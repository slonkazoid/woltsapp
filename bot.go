package main

import (
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func Bot(botLog waLog.Logger, loginLog waLog.Logger, clientLog waLog.Logger, qrLog waLog.Logger, db *SqlDB, container *sqlstore.Container, config *Config) {
	client, err := Login(loginLog, clientLog, qrLog, container)
	if err == LoginTimeout {
		botLog.Errorf("%s, retrying after 5 seconds...", err)
		time.Sleep(5 * time.Second)
		return
	} else if err != nil {
		panic(err)
	}

	defer client.RemoveEventHandlers()
	defer client.Disconnect()

	botLog.Infof("initialized")

	cRestart := make(chan struct{})

	client.AddEventHandler(func(evt interface{}) {
		botLog.Debugf("got event %T", evt)
		switch evt.(type) {
		case *events.Message:
			message := evt.(*events.Message)
			go processMessage(message, client, db, config, botLog)
		case *events.LoggedOut:
			botLog.Warnf("logged out, retrying...")
			cRestart <- struct{}{}
		}
	})

	_ = <-cRestart

	botLog.Infof("restarting...")
}

func processMessage(message *events.Message, client *whatsmeow.Client, db *SqlDB, config *Config, botLog waLog.Logger) {
	if message.IsEdit {
		return
	}
	contents := message.Message.GetConversation()
	trimmedContents := strings.TrimSpace(contents)
	if trimmedContents == "" {
		return
	}
	chat := message.Info.Chat.String()
	sender := message.Info.Sender.String()
	botLog.Debugf("got message in %s by %s: %s", chat, sender, contents)

	if !strings.HasPrefix(trimmedContents, config.Prefix) {
		return
	}

	// 0: any message
	// 1: allowed groups
	// 2: admins
	// 3: owner
	permissionLevel := 0

	if message.Info.IsGroup {
		isAllowedGroup, err := db.IsAllowedGroup(chat)
		if err != nil {
			botLog.Errorf("error while handling message: %v", err)
			return
		} else if isAllowedGroup {
			permissionLevel = 1
		}
	}

	// TODO: admin users

	if message.Info.IsFromMe {
		permissionLevel = 3
	}

	// don't care about random messages atm
	if permissionLevel < 1 {
		return
	}

	commandId := Ulid()
	cmdLog := botLog.Sub(commandId.String())

	body := trimmedContents[len(config.Prefix):]
	cmdLog.Debugf("body: %#v", body)
	argv := MapTrimFilter(strings.Split(body, " "))
	argc := len(argv)
	if argc == 0 {
		return
	}

	cmdLog.Debugf("before lookup: %#v", argv[0])
	argv[0] = strings.ToLower(I18nLookupCommand(argv[0]))
	cmdLog.Infof("command %#v by %s", argv[0], sender)
	cmdLog.Debugf("argc %d argv %#v", argc, argv)
	exec := Commands[argv[0]]
	if exec == nil {
		cmdLog.Errorf("command not found")
		return
	}
	err := exec(argv, argc, permissionLevel, message, client, db, config, cmdLog)
	if err != nil {
		cmdLog.Errorf("error while executing command: %v", err)
		_, err := Reply(client, message, I18nFormat("commandError", err))
		if err != nil {
			cmdLog.Errorf("error while reporting error: %v", err)
		}
	}
}
