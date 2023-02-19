package models

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

type Task struct {
	TaskName string
	DueDate  string
}

type DecoratedRow struct {
	*sql.Row
}

func (r *DecoratedRow) DecorateScan(retryAttempts int, retryDelay int, backoff float64, dest ...any) error {
	var err error
	for i := 0; i < retryAttempts; i++ {
		err = r.Scan(dest...)
		if err == nil || err == sql.ErrNoRows {
			return nil
		} else {
			time.Sleep(time.Duration(float64(retryDelay)*math.Pow(backoff, float64(i))) * time.Second)
		}
	}
	return err
}

func InsertTask(ctx context.Context, rdbmsSession *sql.DB, task *Task) (error, int) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()
	sqlStatement := `
	INSERT INTO task (task_name, due_date)
	VALUES ($1, $2)
	RETURNING id`
	id := 0
	rows := rdbmsSession.QueryRowContext(ctxTimeout, sqlStatement, task.TaskName, task.DueDate)
	err := (&DecoratedRow{rows}).DecorateScan(3, 1, 1.3, &id)
	if err != nil {
		return err, 0
	}
	log.WithFields(log.Fields{"TaskID": id}).Info("New task created")
	return nil, id
}

func DeleteTask(ctx context.Context, rdbmsSession *sql.DB, taskName string) (error, int) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()
	sqlStatement := `
    DELETE FROM task
    WHERE task_name = $1
    RETURNING id`
	id := 0
	rows := rdbmsSession.QueryRowContext(ctxTimeout, sqlStatement, taskName)
	err := (&DecoratedRow{rows}).DecorateScan(3, 1, 1.3, &id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0
		}
		return err, 0
	}
	log.WithFields(log.Fields{"TaskID": id}).Info("Task deleted")
	return nil, id
}

//goland:noinspection GoUnusedExportedFunction
func GetTask(ctx context.Context, rdbmsSession *sql.DB, taskName string) (*Task, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()
	sqlStatement := `
    SELECT task_name, due_date
    FROM task
    WHERE task_name = $1
    `
	rows, err := rdbmsSession.QueryContext(ctxTimeout, sqlStatement, taskName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warning("Task not found")
			return &Task{}, nil
		}
		log.Error("Some error occurred while querying task:", err)
		return &Task{}, err
	}
	var task Task
	for rows.Next() {
		err = rows.Scan(&task.TaskName, &task.DueDate)
		if err != nil {
			log.Error("Some error occurred while scanning task:", err)
			return &Task{}, err
		}
	}
	return &task, nil
}

func GetAllTasks(ctx context.Context, rdbmsSession *sql.DB) ([]Task, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, ConnectionTimeout*time.Second)
	defer cancel()

	var Tasks []Task

	rows, err := rdbmsSession.QueryContext(ctxTimeout, "SELECT task_name, due_date FROM task ORDER BY due_date DESC;")
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
