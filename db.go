package main

import (
	"database/sql"
)

func (recv *SqlDB) GetPermissionLevel(phone_no string) (int, error) {
	db := (*sql.DB)(recv)
	row := db.QueryRow("SELECT permission_level FROM allowed_users WHERE phone_no=?;")
	var permissionLevel int
	err := row.Scan(&permissionLevel)
	return permissionLevel, err
}
