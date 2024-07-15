package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var I18n func(key string) string

//go:embed i18n
var i18nEmbed embed.FS

func initI18n(lang string) {
	var locale map[string]string
	file, err := i18nEmbed.ReadFile(fmt.Sprintf("i18n/locale.%s.json", lang))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't parse lang file for %s", lang)
		panic(err)
	}
	json.Unmarshal(file, &locale)
	I18n = func(key string) string {
		value, has := locale[key]
		if has {
			return value
		} else {
			return key
		}
	}
}

func bot(mainLog waLog.Logger, loginLog waLog.Logger, clientLog waLog.Logger, qrLog waLog.Logger, container *sqlstore.Container) {
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
		mainLog.Debugf("got event %v", evt)
		switch evt.(type) {
		case events.PermanentDisconnect:
			mainLog.Warnf("permanent disconnect, retrying...")
			cRestart <- struct{}{}
		case events.LoggedOut:
			mainLog.Warnf("logged out, retrying...")
			cRestart <- struct{}{}
		}
	})

	_ = <-cRestart

	mainLog.Infof("restarting...")

	client.RemoveEventHandlers()
	client.Disconnect()
}

func main() {
	logLevel, hasLevel := os.LookupEnv("LOG_LEVEL")
	if !hasLevel {
		logLevel = "INFO"
	}

	lang, hasLang := os.LookupEnv("WOLTSAPP_LANG")
	if !hasLang {
		lang = "en"
	}
	initI18n(lang)

	mainLog := waLog.Stdout("Main", logLevel, true)
	loginLog := waLog.Stdout("Login", logLevel, true)
	dbLog := waLog.Stdout("Database", logLevel, true)
	clientLog := waLog.Stdout("Client", logLevel, true)
	qrLog := waLog.Stdout("QR Server", logLevel, true)

	mainLog.Infof("starting up")

	db, err := sql.Open("sqlite3", "file:store.db?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		panic(err)
	}

	dbLog.Debugf("connected to database")

	container := sqlstore.NewWithDB(db, "sqlite3", dbLog)
	container.Upgrade()

	dbLog.Infof("database ready")

	for {
		bot(mainLog, loginLog, clientLog, qrLog, container)
	}
}
