package main

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

var isMac48 *regexp.Regexp = regexp.MustCompile("^([0-9a-fA-F]{2}[-:]){5}[0-9a-fA-F]{2}$")
var isValidHostname *regexp.Regexp = regexp.MustCompile("^[0-9a-f-]{2,36}$")

func FormatHttpAddr(addr string) (string, error) {
	resolved_addr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return "", err
	}

	var ip string
	if resolved_addr.IP == nil {
		ip = "0.0.0.0"
	} else if resolved_addr.IP.To4() == nil {
		ip = fmt.Sprintf("[%s]", resolved_addr.IP.String())
	} else {
		ip = resolved_addr.IP.String()
	}

	if resolved_addr.Port == 80 {
		return "http://" + ip, nil
	} else {
		return fmt.Sprintf("http://%s:%d", ip, resolved_addr.Port), nil
	}
}

func MapTrimFilter(slice []string) []string {
	newSlice := make([]string, 0, len(slice))
	for _, v := range slice {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			newSlice = append(newSlice, trimmed)
		}
	}
	return newSlice
}

func Ulid() ulid.ULID {
	return ulid.MustNew(uint64(time.Now().UnixMilli()), RandSource)
}

func Reply(client *whatsmeow.Client, message *events.Message, body string) (whatsmeow.SendResponse, error) {
	return client.SendMessage(context.Background(), message.Info.Chat, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(body),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      &message.Info.ID,
				Participant:   proto.String(message.Info.Sender.String()),
				QuotedMessage: message.Message,
			},
		},
	})
}

func IsMac48(addr string) bool {
	return isMac48.Match([]byte(addr))
}

func IsValidHostname(name string) bool {
	return isValidHostname.Match([]byte(name))
}
