package services

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	QueryHashed                 = fmt.Sprintf("SELECT u.password FROM %s u WHERE u.email = ?", tableUsers)
	QueryProfileWithCredentials = fmt.Sprintf("SELECT * FROM %s u JOIN %s r ON u.role_id_fk = r.role_id WHERE u.email = ?", tableUsers, tableRoles)
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *SQLiteRepository) CompareCredentials(credentials Credentials) (*User, error) {
	row := r.db.QueryRow(QueryHashed, credentials.Username)
	var hashed string
	if err := row.Scan(&hashed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("auth failed")
		}
		return nil, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(credentials.Password))
	if err != nil {
		return nil, errors.New("wrong password")
	}
	row = r.db.QueryRow(QueryProfileWithCredentials, credentials.Username)

	var (
		user User
		fk   string
	)
	if err := row.Scan(&user.UserID, &user.Name, &user.Last, &user.Email, &user.Password, &user.PFP, &user.CreatedAt, &fk, &user.Role.RoleID, &user.Role.Name, &user.Role.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	user.Password = ""
	return &user, nil
}
