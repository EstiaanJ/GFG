package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const db_str = "postgresql://postgres:uv7MEyHry3q8v4yQFvpTNV5vgMBL5@51.161.163.66:48338/GFG_Prime?sslmode=disable"

func main() {
	db, err := sql.Open("postgres", db_str)

	env := &Env{db: db}
	if err != nil {
		println("Error opening database:" + err.Error())
	}

	indexHandler(db)

	router := gin.Default()

	router.GET("/accounts", getAccounts)

	// loop through users and create endpoints for each user
	for _, user := range users {
		router.POST(user.endpoint, env.accountUserEndpoint)
	}

	router.POST("/transfer", env.transfer)
	router.POST("/accounts", postAccount)
	router.POST("/api/login", login)

	router.Run("localhost:8384")
}

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
}

func getAccounts(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, 1)
}

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
		println("Error querying database: " + err.Error())
		return
	}

	// Query the database for the account corresponding to the account_id
	var account Account
	err = e.db.QueryRow("SELECT username, account_id FROM accounts WHERE account_id = $1", accountID).Scan(&account.Name, &account.ID)
	if err != nil {
		println("Error querying database: " + err.Error())
		return
	}

	// Get the balance for the account
	balance, err := e.getBalance(accountID)
	if err != nil {
		println("Error querying database: " + err.Error())
		return
	}

	// Add the balance to the account object
	account.Balance = balance

	// Return the account information as JSON
	c.IndentedJSON(http.StatusOK, account)
}

func login(c *gin.Context) {

}

func (e *Env) transfer(c *gin.Context) {
	var transfer Transfer //struct to hold the transfer information
	if err := c.BindJSON(&transfer); err != nil {
		//for debugging purposes print the whole json object to the console
		println("Transfer JSON: " + transfer.From_account_username + " " + transfer.To_account_username + " " + transfer.Amount)
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
	println("balance: " + strconv.FormatFloat(balance, 'f', 2, 64))
	//add transaction to database
	//convert from_account_id and to_account_id to string
	transfer_id := uuid.New()
	from_account_id_str := strconv.Itoa(from_account_id)
	to_account_id_str := strconv.Itoa(to_account_id)
	_, err = e.db.Exec("INSERT INTO transactions (trans_id, amount, from_account_id, to_account_id) VALUES ($1, $2, $3, $4)", transfer_id, transfer.Amount, from_account_id_str, to_account_id_str)
	if err != nil {
		println("Error inserting transaction into database: " + err.Error())
		return
	}
	c.IndentedJSON(http.StatusOK, transfer)
}

func (e *Env) getBalance(accountID int) (float64, error) {
	// Query the database for all transactions related to the account
	rows, err := e.db.Query("SELECT trans_id, amount, from_account_id, to_account_id FROM transactions WHERE from_account_id = $1 OR to_account_id = $1", accountID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var debitSum, creditSum float64

	for rows.Next() {
		var id string
		var fromAccountID, toAccountID int
		var amountStr string
		if err := rows.Scan(&id, &amountStr, &fromAccountID, &toAccountID); err != nil {
			return 0, err
		}

		amountFloat, err := strconv.ParseFloat(strings.Replace(amountStr, "$", "", -1), 64)
		if err != nil {
			return 0, err
		}

		if fromAccountID == accountID {
			debitSum += amountFloat
		}
		if toAccountID == accountID {
			creditSum += amountFloat
		}
	}

	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Calculate the balance based on the debit and credit history
	balance := creditSum - debitSum

	return balance, nil
}

// Structs -----------------------------------------
type User struct {
	Name     string `json:"name"`
	passHash string `json:"passHash"`
	endpoint string `json:"endpoint"`
}

type Account struct {
	Name    string  `json:"name"`
	ID      int     `json:"account_number"`
	Balance float64 `json:"balance"`
}

type AccountRequest struct {
	Username string `json:"username"`
}

type Env struct {
	db *sql.DB
}

// data: {from_acc_name:from_acc_name,to_acc_username:to_acc_username, amount: amount },
type Transfer struct {
	From_account_username string `json:"from_acc_name"`
	To_account_username   string `json:"to_acc_username"`
	Amount                string `json:"amount"`
}

// RAM stuff -----------------------------------------

var users = []User{
	{Name: "Yaqoob", passHash: "c1a20a4f708bd9406f0072d50e3cf4958a11b3d13f9001a585e05ca058f10373", endpoint: "/account/Yaqoob/c1a20a4f708bd9406f0072d50e3cf4958a11b3d13f9001a585e05ca058f10373"},
	{Name: "Kyle", passHash: "2cd01faa99669fe375ef010560ccc13f9604126b8473562533c686d4a5286898", endpoint: "/account/Kyle/2cd01faa99669fe375ef010560ccc13f9604126b8473562533c686d4a5286898"},
	{Name: "Jye", passHash: "44108689ee9588e0d6623ff2c1b71009c63083e674c4a2eed432561655563606", endpoint: "/account/Jye/44108689ee9588e0d6623ff2c1b71009c63083e674c4a2eed432561655563606"},
	{Name: "Almo", passHash: "965d9f902a5f3fcbfaf1c3849842fb389b1f488d48ba769bcbc413a5d4ec4919", endpoint: "/account/Almo/965d9f902a5f3fcbfaf1c3849842fb389b1f488d48ba769bcbc413a5d4ec4919"},
	{Name: "Admin", passHash: "1727b577ad38f5a483a7c3741d46a4f7ab450c0a8553eb1c1db6293fea3c7b97", endpoint: "/account/Almo/1727b577ad38f5a483a7c3741d46a4f7ab450c0a8553eb1c1db6293fea3c7b97"},
}
