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
	return db.Exec("INSERT INTO defined_hosts (name, mac_address) VALUES (?, ?) ON CONFLICT (name) DO UPDATE SET mac_address=excluded.mac_address;", mac, name)
}

func (recv *SqlDB) DeleteHost(name string) (sql.Result, error) {
	db := (*sql.DB)(recv)
	return db.Exec("DELETE FROM defined_hosts WHERE name=?;", name)
}

func (recv *SqlDB) SelectHosts() (map[string]string, error) {
	db := (*sql.DB)(recv)
	rows, err := db.Query("SELECT name, mac_address FROM defined_hosts;")
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]string)
	for rows.Next() {
		var mac_address string
		var name string
		err := rows.Scan(&name, &mac_address)
		if err != nil {
			return nil, err
		}
		hosts[name] = mac_address
	}

	return hosts, nil
}
