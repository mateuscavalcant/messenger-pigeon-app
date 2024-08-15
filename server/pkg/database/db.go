package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// InitializeDB initializes the database connection pool.
func InitializeDB() {
	// Retrieve database credentials from environment variables.
	user := "root"
	password := "2009"
	_db := "mydb"
	dbSource := user + ":" + password + "@tcp(localhost:3306)/" + _db

	// Open a connection to the MySQL database.
	conn, err := sql.Open("mysql", dbSource)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	// Set the maximum number of open connections.
	conn.SetMaxOpenConns(100) // Adjust the value according to your requirements.

	// Ping the database to ensure the connection is successful.
	err = conn.Ping()
	if err != nil {
		log.Fatal("Error pinging the database:", err)
	}

	// Set the package-level `db` variable to the connection.
	db = conn
}

// GetDB returns the connection to the MySQL database.
func GetDB() *sql.DB {
	// Return the existing database connection.
	return db
}
