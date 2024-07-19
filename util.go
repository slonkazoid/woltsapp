package main

import (
	"fmt"
	"net"
)

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
