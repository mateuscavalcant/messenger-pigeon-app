package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// InitializeDB inicializa o pool de conexões com o banco de dados.
func InitializeDB() {
	user := "root"
	password := "2009"
	_db := "mydb"
	dbSource := user + ":" + password + "@tcp(localhost:3306)/" + _db

	// Abre uma conexão com o banco de dados MySQL.
	conn, err := sql.Open("mysql", dbSource)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}

	// Define o número máximo de conexões abertas.
	conn.SetMaxOpenConns(100)

	// Testa a conexão com o banco de dados para garantir que a conexão foi bem-sucedida.
	err = conn.Ping()
	if err != nil {
		log.Fatal("Error pinging the database:", err)
	}

	db = conn
}

// GetDB retorna a conexão com o banco de dados MySQL.
func GetDB() *sql.DB {
	// Retorna a conexão existente com o banco de dados.
	return db
}
