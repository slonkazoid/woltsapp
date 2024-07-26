package main

import (
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type CommandFunc func([]string, int, int, *events.Message, *whatsmeow.Client, *SqlDB, *Config, waLog.Logger) error
type CommandMap map[string]CommandFunc

var Commands CommandMap = CommandMap{
	"help":        help,
	"addgroup":    addGroup,
	"removegroup": removeGroup,
	"wake":        wake,
	"wol":         wake,
	"addhost":     addHost,
	"removehost":  removeHost,
	"hosts":       listHosts,
}

func help(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	_, err := Reply(client, message, I18nFormat("help"))
	return err
}

func addGroup(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	if permissionLevel < 2 {
		logger.Errorf("permission denied")
		_, err := Reply(client, message, I18nFormat("permissionDenied"))
		return err
	}

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
	if permissionLevel < 2 {
		logger.Errorf("permission denied")
		_, err := Reply(client, message, I18nFormat("permissionDenied"))
		return err
	}

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
			_, err := Reply(client, message, fmt.Sprintf("%s: %#v\n(%s)", I18nFormat("unknownHost"), argv[1], I18nFormat("mistypeMac")))
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

func addHost(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	if argc == 2 {
		logger.Errorf("mac not specified")
		_, err := Reply(client, message, I18nFormat("macUnspecified"))
		return err
	} else if argc < 3 {
		logger.Errorf("hostname not specified")
		_, err := Reply(client, message, I18nFormat("nameUnspecified"))
		return err
	}

	if !IsValidHostname(argv[1]) {
		logger.Errorf("invalid hostname")
		_, err := Reply(client, message, I18nFormat("nameInvalid", argv[1]))
		return err
	} else if !IsMac48(argv[2]) {
		logger.Errorf("mac not specified")
		_, err := Reply(client, message, I18nFormat("macInvalid", argv[2]))
		return err
	}

	_, err := db.UpsertHost(argv[1], argv[2])
	if err != nil {
		return err
	}

	logger.Infof("host added")
	_, err = Reply(client, message, I18nFormat("hostAdded"))
	return err
}

func removeHost(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	if argc < 2 {
		logger.Errorf("hostname not specified")
		_, err := Reply(client, message, I18nFormat("nameUnspecified"))
		return err
	}

	res, err := db.DeleteHost(argv[1])
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		logger.Warnf("hostname not found")
		_, err = Reply(client, message, I18nFormat("unknownHost"))
		return err
	}

	logger.Infof("host removed")
	_, err = Reply(client, message, I18nFormat("hostRemoved"))
	return err
}

func listHosts(argv []string, argc int, permissionLevel int, message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, logger waLog.Logger) error {
	res, err := db.SelectHosts()
	if err != nil {
		return err
	}

	str := ""
	for k, v := range res {
		str += fmt.Sprintf("`%s` *%s*\n", v, k)
	}

	_, err = Reply(client, message, strings.TrimRight(str, "\n"))
	return err
}
