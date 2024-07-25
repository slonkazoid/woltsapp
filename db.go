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

func (recv *SqlDB) LookupHost(name string) (string, bool, error) {
	db := (*sql.DB)(recv)
	row := db.QueryRow("SELECT mac_address FROM defined_hosts WHERE name=?;", name)
	var addr string
	err := row.Scan(&addr)
	if err != nil {
		if err == sql.ErrNoRows {
			return addr, false, nil
		}
		return addr, false, err
	}
	return addr, true, nil

}

func (recv *SqlDB) UpsertHost(mac string, name string) (sql.Result, error) {
	db := (*sql.DB)(recv)
	return db.Exec("UPSERT INTO defined_hosts (name, mac_addres) VALUES (?, ?);", mac, name)
}
