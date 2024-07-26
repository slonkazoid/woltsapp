package main

import "github.com/koding/multiconfig"

type woltsapp struct {
	HTTP_Addr string `default:":8000"`
	Lang      string `default:"en"`
	DB_URL    string `default:"file:store.db?_foreign_keys=on&_journal_mode=WAL"`
	Prefix    string `default:"w,"`
}

type Config woltsapp

func LoadConfig() (*Config, error) {
	m := multiconfig.New()
	config := new(woltsapp)
	err := m.Load(config)
	return (*Config)(config), err
}
