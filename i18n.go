package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

var I18nFormat func(key string, args ...any) string
var I18nLookupCommand func(key string) string

//go:embed i18n
var i18nEmbed embed.FS

var txtRegex *regexp.Regexp = regexp.MustCompile(`^([\w\d-_]+)\.([\w\d-_]+)\.(txt|md)$`)

func InitI18n(lang string) {
	var locale map[string]string
	file, err := i18nEmbed.ReadFile(fmt.Sprintf("i18n/locale.%s.json", lang))
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read locale file for %s", lang)
		panic(err)
	}
	err = json.Unmarshal(file, &locale)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't parse locale file for %s", lang)
		panic(err)
	}

	entries, err := i18nEmbed.ReadDir("i18n")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		info, _ := entry.Info()
		if info.IsDir() {
			continue
		}

		name := info.Name()

		matches := txtRegex.FindSubmatch([]byte(name))
		if matches == nil {
			continue
		}

		if string(matches[2]) == lang {
			body, err := i18nEmbed.ReadFile(path.Join("i18n", name))
			if err != nil {
				fmt.Fprintf(os.Stderr, "couldn't read %s", name)
				continue
			}
			locale[string(matches[1])] = strings.TrimRight(string(body), "\n")
		}
	}

	I18nFormat = func(key string, args ...any) string {
		value, has := locale[key]
		if has {
			return fmt.Sprintf(value, args...)
		} else {
			return key
		}
	}

	var commands map[string]string
	file, err = i18nEmbed.ReadFile(fmt.Sprintf("i18n/commands.%s.json", lang))
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't read commands file for %s", lang)
		panic(err)
	}
	err = json.Unmarshal(file, &commands)
	if err != nil {
		fmt.Fprintf(os.Stderr, "couldn't parse commands file for %s", lang)
		panic(err)
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
