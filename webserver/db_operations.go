package webserver

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func Connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"))
	db, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		return nil, err
	}
	log.Println("INFO: Connected to PostgresSQL database")
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
