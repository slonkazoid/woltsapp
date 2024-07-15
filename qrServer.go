package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
)

//go:embed static
var staticEmbed embed.FS

//go:embed templates
var tepmlatesEmbed embed.FS

var currentQr string

func listenToQrChanges(cQr chan string) {
	for {
		val := <-cQr
		if val == "" {
			return
		}
		currentQr = val
	}
}

func QrServer(qr string, cQr chan string, cSrv chan *http.Server) {
	currentQr = qr
	go listenToQrChanges(cQr)

	mux := http.NewServeMux()
	addr, hasAddr := os.LookupEnv("WOLTSAPP_HTTP_ADDR")
	if !hasAddr {
		addr = ":8000"
	}
	server := &http.Server{Addr: addr, Handler: mux}

	indexTemplate, err := template.New("index.html").Funcs(template.FuncMap{
		"i18n": I18n,
	}).ParseFS(tepmlatesEmbed, "templates/index.html")
	if err != nil {
		panic(err)
	}

	staticRoot, _ := fs.Sub(staticEmbed, "static")
	staticFs := http.FS(staticRoot)
	mux.Handle("/", http.FileServer(staticFs))
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		indexTemplate.Execute(w, nil)
	})

	mux.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, currentQr)
	})

	cSrv <- server

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		panic(err)
	}
}
