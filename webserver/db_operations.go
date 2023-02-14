package webserver

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	Host     = "aws-postgres.cppudhyhknsc.eu-north-1.rds.amazonaws.com"
	Port     = 5432
	User     = "postgres"
	Password = "password123"
	DbName   = "web_server"
)

func Connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		Host, Port, User, Password, DbName)
	db, err := sql.Open("pgx", psqlInfo)
	log.Println("Connected to PostgresSQL database")
	return db, err
}

func CreateTableTask(db *sql.DB) error {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS task (id SERIAL PRIMARY KEY, task_name TEXT, due_date TIMESTAMPTZ);")
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}

func CreateTableUsers(db *sql.DB) error {
	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, user_name TEXT UNIQUE, password TEXT);")
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}
