package main

import (
	"github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type CommandFunc func([]string, int, int, *events.Message, *whatsmeow.Client, *SqlDB, *Config, waLog.Logger) error
type CommandMap map[string]CommandFunc

var Commands CommandMap = CommandMap{
	"addGroup":    addGroup,
	"removeGroup": removeGroup,
}

func addGroup(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	_, err := db.InsertGroup(message.Info.Chat.String())

	sqliteError, ok := err.(sqlite3.Error)
	if ok && sqliteError.ExtendedCode == sqlite3.ErrConstraintUnique {
		logger.Warnf("group already added")
		_, err := Reply(client, message, I18nFormat("groupExists"))
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	logger.Infof("group added")
	_, err = Reply(client, message, I18nFormat("groupAdded"))
	return err
}

func removeGroup(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	res, err := db.DeleteGroup(message.Info.Chat.String())
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		logger.Warnf("group not found")
		_, err = Reply(client, message, I18nFormat("groupNotFound"))
		return err
	}

	logger.Infof("group removed")
	_, err = Reply(client, message, I18nFormat("groupRemoved"))
	return err
}
