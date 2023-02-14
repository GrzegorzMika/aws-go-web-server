package models

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type AppUser struct {
	UserName string
	Password string
}

func GetUser(rdbmsSession *sql.DB, userName string) (*AppUser, error) {
	rows, err := rdbmsSession.Query("SELECT user_name, password FROM users WHERE user_name = $1 LIMIT 1;", userName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return &AppUser{}, nil
		}
		log.Println("Some error occurred while querying user:", err)
		return &AppUser{}, err
	}
	var user AppUser
	for rows.Next() {
		err = rows.Scan(&user.UserName, &user.Password)
		if err != nil {
			log.Println("Some error occurred while scanning user:", err)
			return &AppUser{}, err
		}
	}
	return &user, nil
}

func InsertUser(rdbmsSession *sql.DB, user *AppUser) (error, int) {
	sqlStatement := `
	INSERT INTO users (user_name, password)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	err := rdbmsSession.QueryRow(sqlStatement, user.UserName, string(password)).Scan(&id)
	if err != nil {
		return err, 0
	}
	log.Println("New record ID is:", id)
	return nil, id
}

func DeleteUser(rdbmsSession *sql.DB, userName string) (error, int) {
	sqlStatement := `
    DELETE FROM users
    WHERE user_name = $1
    RETURNING id`
	id := 0
	err := rdbmsSession.QueryRow(sqlStatement, userName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.Println("Deleted record ID is:", id)
	return nil, id
}
