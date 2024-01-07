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
	Pssword  string `json:"password"`
}

func (r *SQLiteRepository) CompareCredentials(credentials Credentials) (*User, *Role, error) {
	row := r.db.QueryRow(QueryCredentials, credentials.Username, credentials.Pssword)

	var user User
	var role Role
	if err := row.Scan(&user.UserID, &user.Name, &user.Email, &user.Password, &user.PFP, &user.CreatedAt, &user.RoleID, &role.RoleID, &role.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrNotExists
		}
		return nil, nil, err
	}
	return &user, &role, nil
}
