package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mau.fi/whatsmeow"
	//"go.mau.fi/whatsmeow/types/events"
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

	mainLog.Infof("starting up")

	var client *whatsmeow.Client
	var err error

loginStart:
	client, err = Login(loginLog, dbLog, clientLog)
	if err != nil {
		mainLog.Errorf("couldn't log in: %s, retrying after 5 seconds...", err)
		time.Sleep(5 * time.Second)
		goto loginStart
	}

	mainLog.Infof("initialized")

	client.AddEventHandler(func(evt interface{}) {
		fmt.Println("got event:", evt)
	})

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	client.Disconnect()
}
