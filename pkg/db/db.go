package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title VARCHAR(255) NOT NULL,
	comment TEXT,
	repeat VARCHAR(128)
);

CREATE INDEX IF NOT EXISTS 
	scheduler_date ON scheduler(date);
`

func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := err != nil

	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	if install {
		if _, err = db.Exec(schema); err != nil {
			return err
		}
	}

	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func SetDB(database *sql.DB) {
	db = database
}

func AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	return id, err
}

func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat
			FROM scheduler
			WHERE id = ?`
	row := db.QueryRow(query, id)

	var task Task
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func UpdateTask(task *Task) error {
	query := `UPDATE scheduler 
			  SET date=?, title=?, comment=?, repeat=? 
			  WHERE id=?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id=?`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	_, err := db.Exec(query, next, id)
	return err
}

func Tasks(limit int) ([]*Task, error) {
	query := `SELECT id, date, title, comment, repeat
			FROM scheduler
			ORDER BY date ASC
			LIMIT ?`
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	if tasks == nil {
		tasks = []*Task{}
	}

	return tasks, rows.Err()
}
