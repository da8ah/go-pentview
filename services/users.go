package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

const tableUsers = "users"

var (
	QueryCreateUser     = fmt.Sprintf("INSERT INTO %s(name, email, pfp, createdAt, role_id_fk) values(?,?,?,?,?)", tableUsers)
	QueryReadUser       = fmt.Sprintf("SELECT * FROM %s", tableUsers)
	QueryReadUserByName = fmt.Sprintf("SELECT * FROM %s WHERE name = ?", tableUsers)
	QueryUpdateUser     = fmt.Sprintf("UPDATE %s SET name = ?, email = ?, pfp = ?, createdAt = ?, role_id_fk = ? WHERE id = ?", tableUsers)
	QueryDeleteUser     = fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableUsers)
)

type User struct {
	UserID    int64
	Name      string
	Email     string
	PFP       string
	CreatedAt string
	RoleID    int64
}

func (r *SQLiteRepository) CreateUser(user User) (*User, error) {
	res, err := r.db.Exec(QueryCreateUser, user.Name, user.Email, user.PFP, user.CreatedAt, user.RoleID)
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
	user.UserID = id

	return &user, nil
}

func (r *SQLiteRepository) AllUsers() ([]User, error) {
	rows, err := r.db.Query(QueryReadUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.Name, &user.Email, &user.PFP, &user.CreatedAt, &user.RoleID); err != nil {
			return nil, err
		}
		all = append(all, user)
	}
	return all, nil
}

func (r *SQLiteRepository) GetUserByName(name string) (*User, error) {
	row := r.db.QueryRow(QueryReadUserByName, name)

	var user User
	if err := row.Scan(&user.UserID, &user.Name, &user.Email, &user.PFP, &user.CreatedAt, &user.RoleID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &user, nil
}

func (r *SQLiteRepository) UpdateUser(id int64, updated User) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec(QueryUpdateUser, updated.Name, updated.Email, updated.PFP, updated.CreatedAt, updated.RoleID, id)
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

func (r *SQLiteRepository) DeleteUser(id int64) error {
	res, err := r.db.Exec(QueryDeleteUser, id)
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
