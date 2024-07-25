package main

import (
	"fmt"

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
	"wake":        wake,
	"wol":         wake,
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

func wake(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	if argc < 2 {
		logger.Errorf("host not specified")
		_, err := Reply(client, message, I18nFormat("hostUnspecified"))
		return err
	}

	var addr string
	if IsMac48(argv[1]) {
		addr = argv[1]
	} else {
		found, has, err := db.LookupHost(argv[1])
		if err != nil {
			return err
		} else if !has {
			logger.Errorf("unknown hostname: %#v", argv[1])
			_, err := Reply(client, message, fmt.Sprintf("%s: %#v\n%s", I18nFormat("unknownHost"), argv[1], I18nFormat("mistypeMac")))
			return err
		}
		addr = found
	}

	err := WakeByMacString(addr)
	if err != nil {
		return err
	}

	logger.Infof("wakeup sent")
	_, err = Reply(client, message, I18nFormat("wolSent"))
	return err
}
