package main

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func Bot(mainLog waLog.Logger, loginLog waLog.Logger, clientLog waLog.Logger, qrLog waLog.Logger, container *sqlstore.Container, config *Config) {
	client, err := Login(loginLog, clientLog, qrLog, container)
	if err == LoginTimeout {
		mainLog.Errorf("%s, retrying after 5 seconds...", err)
		time.Sleep(5 * time.Second)
		return
	} else if err != nil {
		panic(err)
	}

	defer client.RemoveEventHandlers()
	defer client.Disconnect()

	mainLog.Infof("initialized")

	cRestart := make(chan struct{})

	client.AddEventHandler(func(evt interface{}) {
		mainLog.Debugf("got event %T", evt)
		switch evt.(type) {
		case *events.Message:
			message := evt.(*events.Message)
			if message.IsEdit {
				return
			}
			contents := message.Message.GetConversation()
			if contents == "" {
				return
			}
			chat := message.Info.Chat.String()
			mainLog.Debugf("got message in %s: %s", chat, contents)
			if chat == config.Group_ID {
				mainLog.Infof("got message in bot group: %s", contents)
			}
		case *events.LoggedOut:
			mainLog.Warnf("logged out, retrying...")
			cRestart <- struct{}{}
		}
	})

	_ = <-cRestart

	mainLog.Infof("restarting...")
}
