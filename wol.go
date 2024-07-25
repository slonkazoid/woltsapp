package main

import (
	"errors"
	"net"
)

type WoLPacket = [102]byte

var header []byte = []byte{255, 255, 255, 255, 255, 255}

var InvalidMAC error = errors.New("invalid IEEE 802 MAC-48 address")

func MagicPacket(addr net.HardwareAddr) (packet WoLPacket, _ error) {
	if len(addr) != 6 {
		return packet, InvalidMAC
	}

	// Write header
	copy(packet[:], header)

	offset := 6

	for i := 0; i < 16; i++ {
		copy(packet[offset:], addr)
		offset += 6
	}

	return packet, nil
}

func BroadcastWoL(packet WoLPacket) error {
	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		return err
	}
	defer conn.Close()
	i := 0
	for {
		if i == 102 {
			break
		}
		written, err := conn.Write(packet[i:])
		if err != nil {
			return err
		}
		i += written
	}
	return nil
}

func WakeByMacString(mac string) error {
	addr, err := net.ParseMAC(mac)
	if err != nil {
		return err
	}

	packet, err := MagicPacket(addr)
	if err != nil {
		return err
	}

	return BroadcastWoL(packet)
}
