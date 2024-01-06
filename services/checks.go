package services

import (
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

const tableChecks = "checks"

var (
	QueryCreateCheck = fmt.Sprintf("INSERT INTO %s(type, date, user_id_fk) values(?,?,?)", tableChecks)
	QueryReadChecks  = fmt.Sprintf("SELECT * FROM %s", tableChecks)
	QueryUpdateCheck = fmt.Sprintf("UPDATE %s SET type = ?, date = ?, user_id_fk = ? WHERE id = ?", tableChecks)
	QueryDeleteCheck = fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableChecks)
)

type Check struct {
	CheckID int64
	Type    string
	Date    string
	UserID  int64
}

func (r *SQLiteRepository) CreateCheck(check Check) (*Check, error) {
	res, err := r.db.Exec(QueryCreateCheck, check.Type, check.Date, check.UserID)
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
	check.CheckID = id

	return &check, nil
}

func (r *SQLiteRepository) AllChecks() ([]Check, error) {
	rows, err := r.db.Query(QueryReadChecks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Check
	for rows.Next() {
		var check Check
		if err := rows.Scan(&check.CheckID, &check.Type, &check.Date, &check.UserID); err != nil {
			return nil, err
		}
		all = append(all, check)
	}
	return all, nil
}

func (r *SQLiteRepository) UpdateCheck(id int64, updated Check) (*Check, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec(QueryUpdateCheck, updated.Type, updated.Date, updated.UserID, id)
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

func (r *SQLiteRepository) DeleteCheck(id int64) error {
	res, err := r.db.Exec(QueryDeleteCheck, id)
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
