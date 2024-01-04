package services

import "fmt"

const table = "users"

var QueryTable = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        role INTEGER NOT NULL
		);
		`, table)

var (
	QueryCreate     = fmt.Sprintf("INSERT INTO %s(name, email, role) values(?,?,?)", table)
	QueryRead       = fmt.Sprintf("SELECT * FROM %s", table)
	QueryReadByName = fmt.Sprintf("SELECT * FROM %s WHERE name = ?", table)
	QueryUpdate     = fmt.Sprintf("UPDATE %s SET name = ?, email = ?, role = ? WHERE id = ?", table)
	QueryDelete     = fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
)
