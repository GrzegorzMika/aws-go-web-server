package models

import (
	"database/sql"
	"log"
)

type Task struct {
	TaskName string
	DueDate  string
}

func InsertTask(rdbmsSession *sql.DB, task *Task) (error, int) {
	sqlStatement := `
	INSERT INTO task (task_name, due_date)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	err := rdbmsSession.QueryRow(sqlStatement, task.TaskName, task.DueDate).Scan(&id)
	if err != nil {
		return err, 0
	}
	log.Println("New record ID is:", id)
	return nil, id
}

func DeleteTask(rdbmsSession *sql.DB, taskName string) (error, int) {
	sqlStatement := `
    DELETE FROM task
    WHERE task_name = $1
    RETURNING id`
	id := 0
	err := rdbmsSession.QueryRow(sqlStatement, taskName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.Println("Deleted record ID is:", id)
	return nil, id
}
