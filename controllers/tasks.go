package controllers

import (
	"database/sql"
)

type TaskController struct {
	rdbmsSession *sql.DB
}

func NewTaskController(db *sql.DB) *TaskController {
	return &TaskController{
		rdbmsSession: db,
	}
}
