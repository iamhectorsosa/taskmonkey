package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"reflect"
	"time"
)

type status int

const (
	todo status = iota
	inProgress
	done
)

func (s status) String() string {
	return [...]string{"todo", "in progress", "done"}[s]
}

type taskDB struct {
	db      *sql.DB
	dataDir string
}

func (t *taskDB) tableExists() bool {
	if _, err := t.db.Query(`SELECT * FROM tasks`); err == nil {
		return true
	}
	return false
}

func (t *taskDB) createTable() error {
	_, err := t.db.Exec(`CREATE TABLE "tasks" ( "id" INTEGER, "name" TEXT NOT NULL, "project" TEXT, "status" TEXT, "created" DATETIME, PRIMARY KEY("id" AUTOINCREMENT))`)
	return err
}

type task struct {
	ID                    uint
	Name, Project, Status string
	Created               time.Time
}

func (t *taskDB) getTask(id uint) (task, error) {
	var task task
	err := t.db.QueryRow("SELECT * FROM tasks WHERE id = ?", id).Scan(
		&task.ID,
		&task.Name,
		&task.Project,
		&task.Status,
		&task.Created,
	)
	return task, err
}

func (t *taskDB) getTasks() ([]task, error) {
	var tasks []task
	rows, err := t.db.Query("SELECT * FROM tasks ORDER BY created DESC")
	if err != nil {
		return tasks, fmt.Errorf("unable to get values: %w", err)
	}
	for rows.Next() {
		var task task
		err = rows.Scan(
			&task.ID,
			&task.Name,
			&task.Project,
			&task.Status,
			&task.Created,
		)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, task)
	}
	return tasks, err
}

func (t *taskDB) insert(name, project string) error {
	_, err := t.db.Exec("INSERT INTO tasks(name, project, status, created) VALUES(?, ?, ?, ?)", name, project, todo.String(), time.Now())
	return err
}

func (t *taskDB) delete(id uint) error {
	_, err := t.getTask(id)
	if err != nil {
		return err
	}
	_, err = t.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (orig *task) merge(t task) {
	uValues := reflect.ValueOf(&t).Elem()
	oValues := reflect.ValueOf(orig).Elem()

	for i := 0; i < uValues.NumField(); i++ {
		uField := uValues.Field(i).Interface()
		if oValues.CanSet() {
			if v, ok := uField.(int64); ok && uField != 0 {
				oValues.Field(i).SetInt(v)
			}

			if v, ok := uField.(string); ok && uField != "" {
				oValues.Field(i).SetString(v)
			}
		}
	}
}

func (t *taskDB) update(task task) error {
	orig, err := t.getTask(task.ID)
	if err != nil {
		return err
	}
	orig.merge(task)
	_, err = t.db.Exec(
		"UPDATE tasks SET name = ?, project = ?, status = ? WHERE id = ?",
		orig.Name,
		orig.Project,
		orig.Status,
		orig.ID)
	return err
}

func openDB(path string) (*taskDB, error) {
	db, err := sql.Open("sqlite3", filepath.Join(path, "tasks.db"))
	if err != nil {
		return nil, err
	}

	t := taskDB{db, path}

	if !t.tableExists() {
		if err := t.createTable(); err != nil {
			return nil, err
		}
	}
	return &t, nil
}
