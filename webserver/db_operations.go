package webserver

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

const (
	Host     = "aws-postgres.cppudhyhknsc.eu-north-1.rds.amazonaws.com"
	Port     = 5432
	User     = "postgres"
	Password = "password123"
	DbName   = "web_server"
)

type Task struct {
	TaskName string
	DueDate  string
}

type AppUser struct {
	UserName string
	Password string
}

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

func InsertTask(db *sql.DB, t *Task) (error, int) {
	sqlStatement := `
	INSERT INTO task (task_name, due_date)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	err := db.QueryRow(sqlStatement, t.TaskName, t.DueDate).Scan(&id)
	if err != nil {
		return err, 0
	}
	log.Println("New record ID is:", id)
	return nil, id
}

func DeleteTask(db *sql.DB, taskName string) (error, int) {
	sqlStatement := `
    DELETE FROM task
    WHERE task_name = $1
    RETURNING id`
	id := 0
	err := db.QueryRow(sqlStatement, taskName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.Println("Deleted record ID is:", id)
	return nil, id
}

func InsertUser(db *sql.DB, u *AppUser) (error, int) {
	sqlStatement := `
	INSERT INTO users (user_name, password)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	password, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	err := db.QueryRow(sqlStatement, u.UserName, string(password)).Scan(&id)
	if err != nil {
		return err, 0
	}
	log.Println("New record ID is:", id)
	return nil, id
}

func DeleteUser(db *sql.DB, userName string) (error, int) {
	sqlStatement := `
    DELETE FROM users
    WHERE user_name = $1
    RETURNING id`
	id := 0
	err := db.QueryRow(sqlStatement, userName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.Println("Deleted record ID is:", id)
	return nil, id
}

func GetUser(db *sql.DB, userName string) (AppUser, error) {
	rows, err := db.Query("SELECT user_name, password FROM users WHERE user_name = $1 LIMIT 1;", userName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return AppUser{}, nil
		}
		log.Println("Some error occurred while querying user:", err)
		return AppUser{}, err
	}
	var user AppUser
	for rows.Next() {
		err = rows.Scan(&user.UserName, &user.Password)
		if err != nil {
			log.Println("Some error occurred while scanning user:", err)
			return AppUser{}, err
		}
	}
	return user, nil
}
