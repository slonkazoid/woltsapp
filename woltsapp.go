package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/koding/multiconfig"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var I18n func(key string) string

//go:embed i18n
var i18nEmbed embed.FS

type Config struct {
	HTTP_Addr string `default:":8000"`
	Lang      string `default:"en"`
	Group_ID  string
	DB_URL    string `default:"file:store.db?_foreign_keys=on&_journal_mode=WAL"`
}

func LoadConfig() (*Config, error) {
	m := multiconfig.New()
	config := new(Config)
	err := m.Load(config)
	return config, err
}

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

	mainLog := waLog.Stdout("Main", logLevel, true)
	loginLog := waLog.Stdout("Login", logLevel, true)
	dbLog := waLog.Stdout("Database", logLevel, true)
	clientLog := waLog.Stdout("Client", logLevel, true)
	qrLog := waLog.Stdout("QR Server", logLevel, true)

	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	initI18n(config.Lang)

	mainLog.Infof("starting up")

	var db *sql.DB
	db, err = sql.Open("sqlite3", config.DB_URL)
	if err != nil {
		panic(err)
	}

	dbLog.Debugf("connected to database")

	container := sqlstore.NewWithDB(db, "sqlite3", dbLog)
	container.Upgrade()

	dbLog.Infof("database ready")

	for {
		Bot(mainLog, loginLog, clientLog, qrLog, container, config)
	}
}
