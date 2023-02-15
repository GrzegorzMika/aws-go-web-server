package models

import (
	"database/sql"
	"github.com/pkg/errors"
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

func GetTask(rdbmsSession *sql.DB, taskName string) (*Task, error) {
	sqlStatement := `
    SELECT task_name, due_date
    FROM task
    WHERE task_name = $1
    `
	rows, err := rdbmsSession.Query(sqlStatement, taskName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Task not found")
			return &Task{}, nil
		}
		log.Println("Some error occurred while querying task:", err)
		return &Task{}, err
	}
	var task Task
	for rows.Next() {
		err = rows.Scan(&task.TaskName, &task.DueDate)
		if err != nil {
			log.Println("Some error occurred while scanning task:", err)
			return &Task{}, err
		}
	}
	return &task, nil
}

func GetAllTasks(rdbmsSession *sql.DB) ([]Task, error) {
	var Tasks []Task

	rows, err := rdbmsSession.Query("SELECT task_name, due_date FROM task ORDER BY due_date DESC;")
	if err != nil {
		return Tasks, errors.Wrap(err, "Failed to query for all tasks")
	}
	for rows.Next() {
		var task Task
		err = rows.Scan(&task.TaskName, &task.DueDate)
		if err != nil {
			return Tasks, errors.Wrap(err, "Failed to scan task")
		}
		Tasks = append(Tasks, task)
	}
	return Tasks, nil
}
