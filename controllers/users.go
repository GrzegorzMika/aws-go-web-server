package controllers

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
)

type UserController struct {
	rdbmsSession *sql.DB
	redisSession *redis.Client
}

func NewUserController(rdbmsSession *sql.DB, redisSession *redis.Client) *UserController {
	return &UserController{
		rdbmsSession: rdbmsSession,
		redisSession: redisSession,
	}
}
