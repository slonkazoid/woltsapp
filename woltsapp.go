package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type SqlDB sql.DB

var I18nFormat func(key string) string
var I18nLookupCommand func(key string) string

//go:embed i18n
var i18nEmbed embed.FS

//go:embed migrate
var migrateEmbed embed.FS

var RandSource io.Reader

func initI18n(lang string) {
	var locale map[string]string
	file, err := i18nEmbed.ReadFile(fmt.Sprintf("i18n/locale.%s.json", lang))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't read locale file for %s", lang)
		panic(err)
	}
	err = json.Unmarshal(file, &locale)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't parse locale file for %s", lang)
		panic(err)
	}

	var commands map[string]string
	file, err = i18nEmbed.ReadFile(fmt.Sprintf("i18n/commands.%s.json", lang))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't read commands file for %s", lang)
		panic(err)
	}
	err = json.Unmarshal(file, &commands)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "couldn't parse commands file for %s", lang)
		panic(err)
	}

	I18nFormat = func(key string) string {
		value, has := locale[key]
		if has {
			return value
		} else {
			return key
		}
	}
	I18nLookupCommand = func(key string) string {
		value, has := commands[key]
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

	mainLog := waLog.Stdout("Main", logLevel, true)
	botLog := waLog.Stdout("Bot", logLevel, true)
	loginLog := waLog.Stdout("Login", logLevel, true)
	dbLog := waLog.Stdout("Database", logLevel, true)
	clientLog := waLog.Stdout("Client", logLevel, true)
	qrLog := waLog.Stdout("QR Server", logLevel, true)

	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	initI18n(config.Lang)

	RandSource, err = os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	mainLog.Infof("starting up")

	db, err := sql.Open("sqlite3", config.DB_URL)
	if err != nil {
		panic(err)
	}

	dbLog.Debugf("connected to database")

	iodriver, err := iofs.New(migrateEmbed, "migrate")
	if err != nil {
		panic(err)
	}
	dbdriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}
	migrations, err := migrate.NewWithInstance(
		"migrateEmbed", iodriver,
		"store", dbdriver)
	if err != nil {
		panic(err)
	}
	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		panic(err)
	}

	container := sqlstore.NewWithDB(db, "sqlite3", dbLog)
	err = container.Upgrade()
	if err != nil {
		panic(err)
	}

	dbLog.Infof("database up")

	for {
		Bot(botLog, loginLog, clientLog, qrLog, (*SqlDB)(db), container, config)
	}
}
