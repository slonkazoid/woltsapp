package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	waLog "go.mau.fi/whatsmeow/util/log"
	"nhooyr.io/websocket"
)

//go:embed static
var staticEmbed embed.FS

//go:embed templates
var tepmlatesEmbed embed.FS

var currentQr string

/*
ID   MEANING
0    QR update
1    Close
*/
type message struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

func listenToQrChanges(cQr chan string, cClients chan chan string) {
	clients := make([]chan string, 0)
forever:
	for {
		select {
		case val := <-cQr:
			if val == "" {
				break forever
			}
			currentQr = val
			for _, client := range clients {
				client <- val
			}
		case newClient := <-cClients:
			if newClient == nil {
				break forever
			}
			clients = append(clients, newClient)
		}
	}
	for _, client := range clients {
		close(client)
	}
}

func QrServer(log waLog.Logger, qr string, cQr chan string, cSrv chan *http.Server) {
	currentQr = qr
	cClients := make(chan chan string, 1)
	go listenToQrChanges(cQr, cClients)

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
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
		if err != nil {
			log.Errorf("%v", err)
			return
		}
		defer c.CloseNow()

		log.Debugf("got new client %s", r.RemoteAddr)

		jsonBody, jsonErr := json.Marshal(message{
			Id:   0,
			Body: currentQr,
		})
		if jsonErr != nil {
			panic(jsonErr)
		}
		c.Write(r.Context(), websocket.MessageText, jsonBody)

		thisChannel := make(chan string)
		cClients <- thisChannel

		log.Debugf("added to broadcast queue")

		for {
			val := <-thisChannel
			if val == "" {
				jsonBody, _ = json.Marshal(message{
					Id:   1,
					Body: "",
				})
				c.Write(r.Context(), websocket.MessageText, jsonBody)
				return
			}
			jsonBody, _ = json.Marshal(message{
				Id:   0,
				Body: val,
			})
			c.Write(r.Context(), websocket.MessageText, jsonBody)
		}
	})

	cSrv <- server

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		panic(err)
	}
}
