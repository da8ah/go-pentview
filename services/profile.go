package services

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	QueryReadProfile   = fmt.Sprintf("SELECT * FROM %s WHERE user_id = ?", tableUsers)
	QueryUpdateProfile = fmt.Sprintf("UPDATE %s SET name = ?, last = ?, email = ? WHERE user_id = ?", tableUsers)
)

func (r *SQLiteRepository) GetProfileById(id int64) (*User, error) {
	row := r.db.QueryRow(QueryReadProfile, id)

	var user User
	if err := row.Scan(&user.UserID, &user.Name, &user.Last, &user.Email, &user.Password, &user.PFP, &user.CreatedAt, &user.RoleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

func (r *SQLiteRepository) UpdateProfile(id int64, updated User) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec(QueryUpdateProfile, updated.Name, updated.Last, updated.Email, id)
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
