package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func postAccount(c *gin.Context) {
	var newAccount Account
	if err := c.BindJSON(&newAccount); err != nil {
		return
	}
	//accounts = append(accounts, newAccount)
	c.IndentedJSON(http.StatusCreated, newAccount)
}

func (e *Env) accountUserEndpoint(c *gin.Context) {
	var req AccountRequest
	if err := c.BindJSON(&req); err != nil {
		println("Couldn't bind username")
		return
	}

	// Query the database for the account_id corresponding to the given username
	var accountID int
	err := e.db.QueryRow("SELECT account_id FROM users WHERE username = $1", req.Username).Scan(&accountID)
	if err != nil {
		println("[App Server - Get acc-id from DB] Error querying database: " + err.Error())
		return
	}

	// Query the database for the account corresponding to the account_id
	var account Account
	err = e.db.QueryRow("SELECT username, account_id FROM accounts WHERE account_id = $1", accountID).Scan(&account.Name, &account.ID)
	if err != nil {
		println("[App Server - Get account from DB] Error querying database: " + err.Error())
		return
	}

	// Get the balance for the account
	balance, err := e.getBalance(accountID)
	if err != nil {
		println("[App Server - Determine balance] Error querying database: " + err.Error())
		return
	}

	//Get Last 5 transactions
	var transactions []Transaction
	rows, err := e.db.Query("SELECT * FROM transactions WHERE from_account_id = $1 OR to_account_id = $1 ORDER BY date DESC LIMIT 5", accountID)
	if err != nil {
		println("[App Server - Get transactions from DB] Error querying database: " + err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var transaction Transaction
		//
		//Scan all 6 columns like this err = rows.Scan(&transaction.Amount, &transaction.Date, &transaction.GameDate)
		//Amount          float64 `json:"amount"` Date            string  `json:"date"` GameDate        string  `json:"game_date"` Trans_id        int     `json:"trans_id"` From_Account_id int     `json:"from_account_id"` To_Account_id   int     `json:"to_account_id"`
		//err = rows.Scan(&transaction.Amount, &transaction.Date, &transaction.GameDate, &transaction.Trans_id, &transaction.From_Account_id, &transaction.To_Account_id)
		//Order of colums is trans_id, from_account_id, to_account_id, game_date, date, amount
		err = rows.Scan(&transaction.Trans_id, &transaction.From_Account_id, &transaction.To_Account_id, &transaction.GameDate, &transaction.Date, &transaction.Amount, &transaction.Description)

		if err != nil {
			println("[App Server - Get transaction history from DB] Error scanning row: " + err.Error())
			continue
		}
		transactions = append(transactions, transaction)
	}

	account.Transactions = transactions

	// Add the balance to the account object
	account.Balance = balance

	// Return the account information as JSON
	//Add Access-Control-Allow-Origin header to the response
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, account)
}

func (e *Env) transfer(c *gin.Context) {
	var transfer Transfer //struct to hold the transfer information
	if err := c.BindJSON(&transfer); err != nil {
		//for debugging purposes print the whole json object to the console
		println("Transfer JSON: " + transfer.From_account_username + " " + transfer.To_account_username + " " + transfer.Amount + " " + transfer.Description)
		println("Couldn't bind transfer")
		return
	}

	//get to_account_id from database
	var to_account_id int
	println("toacc: " + transfer.To_account_username)
	err2 := e.db.QueryRow("SELECT account_id FROM accounts WHERE username = $1", transfer.To_account_username).Scan(&to_account_id)
	if err2 != nil {
		println("Error querying database for the to account: " + err2.Error())
		return
	}

	//get from_account_id from database
	var from_account_id int
	err := e.db.QueryRow("SELECT account_id FROM accounts WHERE username = $1", transfer.From_account_username).Scan(&from_account_id)
	if err != nil {
		println("Error querying database for username for the from account: " + err.Error())
		return
	}

	println("fromacc: " + transfer.From_account_username + " " + strconv.Itoa(from_account_id))
	//assign balance to variable
	balance, err := e.getBalance(from_account_id)
	if err != nil {
		println("Error querying database for determining balance: " + err.Error())
		return
	}
	//convert transfer amount to float
	amountFloat, err := strconv.ParseFloat(strings.Replace(transfer.Amount, "$", "", -1), 64)
	if err != nil {
		println("Error converting transfer amount to float: " + err.Error())
		return
	}
	if balance < amountFloat {
		println("Insufficient funds")
		return
	}
	if amountFloat < 0 {
		println("Cannot transfer negative amount")
		return
	}

	println("balance: " + strconv.FormatFloat(balance, 'f', 2, 64))
	//add transaction to database
	//convert from_account_id and to_account_id to string
	transfer_id := uuid.New()
	from_account_id_str := strconv.Itoa(from_account_id)
	to_account_id_str := strconv.Itoa(to_account_id)
	_, err = e.db.Exec("INSERT INTO transactions (trans_id, amount, from_account_id, to_account_id, date, game_date, description) VALUES ($1, $2, $3, $4, $5, $6, $7)", transfer_id, transfer.Amount, from_account_id_str, to_account_id_str, time.Now().Format("2006-01-02 15:04:05"), "33.123", transfer.Description)
	if err != nil {
		println("Error inserting transaction into database: " + err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
	c.IndentedJSON(http.StatusOK, transfer)
}

func getAccounts(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, 1)
}

func (e *Env) getBalance(accountID int) (float64, error) {
	var balance float64
	err := e.db.QueryRow("SELECT SUM(CASE WHEN from_account_id = $1 THEN -amount ELSE amount END) FROM transactions WHERE from_account_id = $1 OR to_account_id = $1", accountID).Scan(&balance)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

/*
func indexHandler(db *sql.DB) {
	var username string
	var usernames []string
	//query database for the usernames in the accounts table
	rows, err := db.Query("SELECT username FROM accounts")
	defer rows.Close()
	if err != nil {
		println("Error querying database:" + err.Error())
	}
	for rows.Next() {
		rows.Scan(&username)
		usernames = append(usernames, username)
	}
	//loop and print usernames
	for _, username := range usernames {
		println(username)
	}
} */
