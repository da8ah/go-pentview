package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

const tableClockings = "clockings"

var (
	QueryCreateClocking   = fmt.Sprintf("INSERT INTO %s(type, date, user_id_fk) values(?,?,?)", tableClockings)
	QueryReadClockings    = fmt.Sprintf("SELECT * FROM %s", tableClockings)
	QueryReadClockingById = fmt.Sprintf("SELECT * FROM %s WHERE clocking_id = ?", tableClockings)
	QueryUpdateClocking   = fmt.Sprintf("UPDATE %s SET type = ?, date = ?, user_id_fk = ? WHERE id = ?", tableClockings)
	QueryDeleteClocking   = fmt.Sprintf("DELETE FROM %s WHERE clocking_id = ?", tableClockings)
)

type Clocking struct {
	ClockingID int64  `json:"_id"`
	Type       string `json:"type"`
	Date       string `json:"date"`
	UserID     int64  `json:"user"`
}

func (r *SQLiteRepository) CreateClocking(clocking Clocking) (*Clocking, error) {
	res, err := r.db.Exec(QueryCreateClocking, clocking.Type, clocking.Date, clocking.UserID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	clocking.ClockingID = id

	return &clocking, nil
}

func (r *SQLiteRepository) AllClockings() ([]Clocking, error) {
	rows, err := r.db.Query(QueryReadClockings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Clocking
	for rows.Next() {
		var clocking Clocking
		if err := rows.Scan(&clocking.ClockingID, &clocking.Type, &clocking.Date, &clocking.UserID); err != nil {
			return nil, err
		}
		all = append(all, clocking)
	}
	return all, nil
}

func (r *SQLiteRepository) GetClockingById(id int64) (*Clocking, error) {
	row := r.db.QueryRow(QueryReadClockingById, id)

	var clocking Clocking
	if err := row.Scan(&clocking.ClockingID, &clocking.Type, &clocking.Date, &clocking.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &clocking, nil
}

func (r *SQLiteRepository) UpdateClocking(id int64, updated Clocking) (*Clocking, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec(QueryUpdateClocking, updated.Type, updated.Date, updated.UserID, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) DeleteClocking(id int64) error {
	res, err := r.db.Exec(QueryDeleteClocking, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}
