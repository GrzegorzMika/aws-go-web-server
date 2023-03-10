package models

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AppUser struct {
	UserName string
	Password string
}

func GetUser(ctx context.Context, rdbmsSession *sql.DB, userName string) (*AppUser, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()
	rows, err := rdbmsSession.QueryContext(ctxTimeout, "SELECT user_name, password FROM users WHERE user_name = $1 LIMIT 1;", userName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warning("User not found")
			return &AppUser{}, nil
		}
		log.Error("Some error occurred while querying user:", err)
		return &AppUser{}, err
	}
	var user AppUser
	for rows.Next() {
		err = rows.Scan(&user.UserName, &user.Password)
		if err != nil {
			log.Error("Some error occurred while scanning user:", err)
			return &AppUser{}, err
		}
	}
	return &user, nil
}

//goland:noinspection GoUnusedExportedFunction
func InsertUser(ctx context.Context, rdbmsSession *sql.DB, user *AppUser) (error, int) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()
	sqlStatement := `
	INSERT INTO users (user_name, password)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	err := rdbmsSession.QueryRowContext(ctxTimeout, sqlStatement, user.UserName, string(password)).Scan(&id)
	if err != nil {
		return err, 0
	}
	log.Info("New user created with ID:", id)
	return nil, id
}

//goland:noinspection GoUnusedExportedFunction
func DeleteUser(ctx context.Context, rdbmsSession *sql.DB, userName string) (error, int) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()
	sqlStatement := `
    DELETE FROM users
    WHERE user_name = $1
    RETURNING id`
	id := 0
	err := rdbmsSession.QueryRowContext(ctxTimeout, sqlStatement, userName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.Info("User deleted successfully with id:", id)
	return nil, id
}
