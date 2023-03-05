package main

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

//The Postgres URI and password that is leaked in the version history here has been changed

func main() {
	argsWithoutProg := os.Args[1:]

	serv_ip := argsWithoutProg[0]
	serv_port := argsWithoutProg[1]
	db_ip := argsWithoutProg[2]
	db_port := argsWithoutProg[3]
	db_name := argsWithoutProg[4]
	postgresURI := argsWithoutProg[5] + "@" + db_ip + ":" + db_port + "/" + db_name + "?sslmode=disable"

	ipAndPort := "http://" + db_ip + ":" + db_port

	println("Launching on " + ipAndPort)

	db, err := sql.Open("postgres", postgresURI)

	env := &Env{db: db}
	if err != nil {
		println("Error opening database:" + err.Error())
	}

	//indexHandler(db)

	//Get all users from the database for the endpoints
	rows, err := db.Query("SELECT username, pass_hash FROM users")
	defer rows.Close()
	if err != nil {
		println("Error querying database:" + err.Error())
	}

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Name, &user.passHash)
		if err != nil {
			println("Error scanning row: " + err.Error())
			continue
		}
		user.endpoint = "/account/" + user.Name + "/" + user.passHash
		users = append(users, user)
		//print user passHash
		println("passHash: " + user.passHash)
	}

	router := gin.Default()

	router.GET("/accounts", getAccounts)

	//Print number of users
	println("Number of users: " + strconv.Itoa(len(users)))
	// loop through users and create endpoints for each user
	for _, user := range users {

		//Print the endpoint
		println("Endpoint: " + user.endpoint)
		router.POST(user.endpoint, env.accountUserEndpoint)

	}

	router.POST("/transfer", env.transfer)
	router.POST("/accounts", postAccount)
	router.POST("/api/login", login)

	router.Use(cors.Default())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{ipAndPort},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin"},
	}))

	router.Run(serv_ip + ":" + serv_port)
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

	// Add the balance to the account object
	account.Balance = balance

	// Return the account information as JSON
	//Add Access-Control-Allow-Origin header to the response
	c.Header("Access-Control-Allow-Origin", "*")
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
	_, err = e.db.Exec("INSERT INTO transactions (trans_id, amount, from_account_id, to_account_id, date) VALUES ($1, $2, $3, $4, $5)", transfer_id, transfer.Amount, from_account_id_str, to_account_id_str, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		println("Error inserting transaction into database: " + err.Error())
		return
	}
	c.Header("Access-Control-Allow-Origin", "*")
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
