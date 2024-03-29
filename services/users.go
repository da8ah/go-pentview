package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const tableUsers = "users"

var (
	QueryCreateUser   = fmt.Sprintf("INSERT INTO %s(name, last, email, password, pfp, createdAt, role_id_fk) values(?,?,?,?,?,?,?)", tableUsers)
	QueryReadUser     = fmt.Sprintf("SELECT * FROM %s u JOIN %s r ON u.role_id_fk = r.role_id", tableUsers, tableRoles)
	QueryReadUserById = fmt.Sprintf("SELECT * FROM %s u JOIN %s r ON u.role_id_fk = r.role_id WHERE user_id = ?", tableUsers, tableRoles)
	QueryUpdateUser   = fmt.Sprintf("UPDATE %s SET name = ?, last = ?, email = ? WHERE user_id = ?", tableUsers)
	QueryDeleteUser   = fmt.Sprintf("DELETE FROM %s WHERE user_id = ?", tableUsers)
)

type User struct {
	UserID    int64  `json:"_id"`
	Name      string `json:"firstName"`
	Last      string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	PFP       string `json:"profileImage"`
	CreatedAt string `json:"createdAt"`
	Role      Role   `json:"role"`
}
type UserToCreate struct {
	UserID    int64  `json:"_id"`
	Name      string `json:"firstName"`
	Last      string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	PFP       string `json:"profileImage"`
	CreatedAt string `json:"createdAt"`
	Role      string `json:"role"`
}

func (r *SQLiteRepository) CreateUser(userToCreate UserToCreate) error {
	role_id, err := strconv.ParseInt(userToCreate.Role, 10, 64)
	if err != nil {
		return errors.New("role id must be string")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(userToCreate.Password), bcrypt.DefaultCost)
	userToCreate.Password = string(hashed)
	_, err = r.db.Exec(QueryCreateUser, userToCreate.Name, userToCreate.Last, userToCreate.Email, userToCreate.Password, userToCreate.PFP, time.Now().Format(time.RFC3339), role_id)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return ErrDuplicate
			}
		}
		return err
	}

	return nil
}

func (r *SQLiteRepository) AllUsers() ([]User, error) {
	rows, err := r.db.Query(QueryReadUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []User
	for rows.Next() {
		var (
			user User
			fk   string
		)
		if err := rows.Scan(&user.UserID, &user.Name, &user.Last, &user.Email, &user.Password, &user.PFP, &user.CreatedAt, &fk, &user.Role.RoleID, &user.Role.Name, &user.Role.CreatedAt); err != nil {
			return nil, err
		}
		user.Password = ""
		all = append(all, user)
	}
	return all, nil
}

func (r *SQLiteRepository) GetUserById(user_id int64) (*User, error) {
	row := r.db.QueryRow(QueryReadUserById, user_id)

	var user User
	if err := row.Scan(&user.UserID, &user.Name, &user.Last, &user.Email, &user.Password, &user.PFP, &user.CreatedAt, &user.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

func (r *SQLiteRepository) UpdateUser(id int64, updated User) (*User, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec(QueryUpdateUser, updated.Name, updated.Last, updated.Email, id)
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
