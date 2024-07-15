package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

var LoginTimeout error = errors.New("didn't login in time")

// Code from whatsmeow docs
func initWhatsapp(clientLog waLog.Logger, container *sqlstore.Container, cQr chan string, cLoggedIn chan struct{}, cRet chan *whatsmeow.Client) {
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				clientLog.Debugf("QR code event:", evt.Code)
				cQr <- evt.Code
			} else {
				cLoggedIn <- struct{}{}
				clientLog.Debugf("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		cLoggedIn <- struct{}{}
	}

	cRet <- client
}

func Login(loginLog waLog.Logger, clientLog waLog.Logger, qrLog waLog.Logger, container *sqlstore.Container) (*whatsmeow.Client, error) {
	cQr := make(chan string, 1)
	cLoggedIn := make(chan struct{})
	cClient := make(chan *whatsmeow.Client)
	go initWhatsapp(clientLog, container, cQr, cLoggedIn, cClient)

	select {
	case qr := <-cQr:
		loginLog.Debugf("received qr before loggedIn")
		cSrv := make(chan *http.Server, 1)
		go QrServer(qrLog, qr, cQr, cSrv)
		srv := <-cSrv
		addr, _ := net.ResolveTCPAddr("tcp", srv.Addr)
		var ip string
		if addr.IP == nil {
			ip = "0.0.0.0"
		} else if addr.IP.To4() == nil {
			ip = fmt.Sprintf("[%s]", addr.IP.String())
		} else {
			ip = addr.IP.String()
		}
		loginLog.Infof("scan QR here: http://%s:%d", ip, addr.Port)
		loginLog.Infof("waiting for QR to be scanned")
		<-cLoggedIn
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()
		srv.Shutdown(shutdownCtx)
		close(cQr)
	case <-cLoggedIn:
		loginLog.Debugf("received loggedIn before qr")
	}

	client := <-cClient

	if !client.IsConnected() {
		return &whatsmeow.Client{}, LoginTimeout
	}

	return client, nil
}
