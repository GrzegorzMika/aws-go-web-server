package controllers

import (
	"context"
	"database/sql"
)

type TaskController struct {
	appContext   context.Context
	rdbmsSession *sql.DB
}

func NewTaskController(ctx context.Context, db *sql.DB) *TaskController {
	return &TaskController{
		appContext:   ctx,
		rdbmsSession: db,
	}
}
