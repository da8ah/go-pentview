package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
)

const tableRoles = "roles"

var (
	QueryCreateRole     = fmt.Sprintf("INSERT INTO %s(name, createdAt) values(UPPER(?),?)", tableRoles)
	QueryReadRoles      = fmt.Sprintf("SELECT * FROM %s", tableRoles)
	QueryReadRoleByName = fmt.Sprintf("SELECT * FROM %s WHERE name = ?", tableRoles)
	QueryUpdateRole     = fmt.Sprintf("UPDATE %s SET name = UPPER(?) WHERE role_id = ?", tableRoles)
	QueryDeleteRole     = fmt.Sprintf("DELETE FROM %s WHERE role_id = ?", tableRoles)
)

type Role struct {
	RoleID    int64  `json:"_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

func (r *SQLiteRepository) CreateRole(role Role) (*Role, error) {
	res, err := r.db.Exec(QueryCreateRole, role.Name, time.Now().Format(time.RFC3339))
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
	role.RoleID = id

	return &role, nil
}

func (r *SQLiteRepository) AllRoles() ([]Role, error) {
	rows, err := r.db.Query(QueryReadRoles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.RoleID, &role.Name, &role.CreatedAt); err != nil {
			return nil, err
		}
		all = append(all, role)
	}
	return all, nil
}

func (r *SQLiteRepository) GetRoleByName(name string) (*Role, error) {
	row := r.db.QueryRow(QueryReadRoleByName, name)

	var role Role
	if err := row.Scan(&role.RoleID, &role.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &role, nil
}

func (r *SQLiteRepository) UpdateRole(id int64, updated Role) (*Role, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec(QueryUpdateRole, updated.Name, id)
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

func (r *SQLiteRepository) DeleteRole(id int64) error {
	res, err := r.db.Exec(QueryDeleteRole, id)
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
