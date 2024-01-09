package services

import (
	"database/sql"
	"errors"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) Migrate() error {
	const QueryTable = `
		CREATE TABLE IF NOT EXISTS roles (
			role_id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			createdAt TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			last TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			pfp TEXT NOT NULL,
			createdAt TEXT NOT NULL,
			role_id_fk INTEGER,
			FOREIGN KEY (role_id_fk)
				REFERENCES roles (role_id)
		);

		CREATE TABLE IF NOT EXISTS clockings (
			clocking_id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			date TEXT NOT NULL,
			user_id_fk INTEGER,
			FOREIGN KEY (user_id_fk)
				REFERENCES users (user_id)
		);`

	_, err := r.db.Exec(QueryTable)
	return err
}
