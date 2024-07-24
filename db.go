package main

import (
	"database/sql"
)

func (recv *SqlDB) IsAllowedGroup(id string) (bool, error) {
	db := (*sql.DB)(recv)
	row := db.QueryRow("SELECT 1 FROM groups WHERE id=?;", id)
	var _int int
	err := row.Scan(&_int)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (recv *SqlDB) InsertGroup(id string) (sql.Result, error) {
	db := (*sql.DB)(recv)
	return db.Exec("INSERT INTO groups (id) VALUES (?);", id)
}

func (recv *SqlDB) DeleteGroup(id string) (sql.Result, error) {
	db := (*sql.DB)(recv)
	return db.Exec("DELETE FROM groups WHERE id=?;", id)
}
