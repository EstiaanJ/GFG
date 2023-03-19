package main

import (
	"database/sql"
	"os"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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

	//ipAndPort := "http://" + db_ip + ":" + db_port

	println("Launching on " + serv_ip + ":" + serv_port)
	println("Postgres URI: " + postgresURI)

	db, err := sql.Open("postgres", postgresURI)

	env := &Env{db: db}
	if err != nil {
		println("Error opening database:" + err.Error())
	}

	//Get all users from the database for the endpoints
	rows, err := db.Query("SELECT username, pass_hash FROM users")
	defer rows.Close()
	if err != nil {
		println("Error querying database:" + err.Error())
	}

	router := gin.Default()

	go setup(router)

	router.GET("/accounts", getAccounts)

	createUsers(rows, err, router, env)

	router.POST("/transfer", env.transfer)
	router.POST("/accounts", postAccount)

	router.Use(cors.Default())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"51.161.163.66", "localhost", "0.0.0.0", "127.0.0.1", "51.161.163.66:44658", "51.161.163.66:48338"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin"},
	}))

	router.Run(serv_ip + ":" + serv_port)
}

func createUsers(rows *sql.Rows, err error, router *gin.Engine, env *Env) []User {
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
		//Print the endpoint
		println("Endpoint: " + user.endpoint)
		router.POST(user.endpoint, env.accountUserEndpoint)
	}
	println("Number of users: " + strconv.Itoa(len(users)))
	return users
}

// Structs -----------------------------------------
type User struct {
	Name     string `json:"name"`
	passHash string `json:"passHash"`
	endpoint string `json:"endpoint"`
}

type Account struct {
	Name         string        `json:"name"`
	ID           int           `json:"account_number"`
	Balance      float64       `json:"balance"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Amount          float64 `json:"amount"`
	Date            string  `json:"date"`
	GameDate        string  `json:"game_date"`
	Trans_id        string  `json:"trans_id"`
	From_Account_id int     `json:"from_account_id"`
	To_Account_id   int     `json:"to_account_id"`
	Description     string  `json:"description"`
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
	Description           string `json:"description"`
}

// RAM stuff -----------------------------------------
