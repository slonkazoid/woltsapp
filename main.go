package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/store/sqlstore"
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
		Bot(mainLog, loginLog, clientLog, qrLog, container)
	}
}
