package main

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func Bot(mainLog waLog.Logger, loginLog waLog.Logger, clientLog waLog.Logger, qrLog waLog.Logger, container *sqlstore.Container) {
	client, err := Login(loginLog, clientLog, qrLog, container)
	if err == LoginTimeout {
		mainLog.Errorf("%s, retrying after 5 seconds...", err)
		time.Sleep(5 * time.Second)
		return
	} else if err != nil {
		panic(err)
	}

	mainLog.Infof("initialized")

	cRestart := make(chan struct{})

	client.AddEventHandler(func(evt interface{}) {
		mainLog.Debugf("got event %T", evt)
		switch evt.(type) {
		case *events.Message:
			message := evt.(*events.Message)
			chat := message.Info.Chat.String()
			mainLog.Infof("got message in %s", chat)
			if chat == "" {
				mainLog.Debugf("got message in bot group: %v", message)
			}
		case *events.LoggedOut:
			mainLog.Warnf("logged out, retrying...")
			cRestart <- struct{}{}
		}
	})

	_ = <-cRestart

	mainLog.Infof("restarting...")

	client.RemoveEventHandlers()
	client.Disconnect()
}
