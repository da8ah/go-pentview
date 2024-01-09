package services

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	QueryCredentials = fmt.Sprintf("SELECT * FROM %s u JOIN %s r ON u.role_id_fk = r.role_id WHERE u.email = ? AND u.password = ?", tableUsers, tableRoles)
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (r *SQLiteRepository) CompareCredentials(credentials Credentials) (*User, error) {
	row := r.db.QueryRow(QueryCredentials, credentials.Username, credentials.Password)

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
