package main

import (
	"log"

	"github.com/tarm/serial"
)

type serialMonitor struct {
	sm *serial.Port

	buffer []byte
	out    []byte
	port   string
}

func newSerialMonitor(port string, baud int) serialMonitor {
	var mon serialMonitor
	mon.buffer = make([]byte, 1)
	mon.port = port

	c := &serial.Config{Name: port, Baud: baud}
	s, _ := serial.OpenPort(c)

	mon.sm = s

	return mon
}

func (s *serialMonitor) readln() []byte {
	s.out = make([]byte, 0)
	for true {
		_, err := s.sm.Read(s.buffer)
		if err != nil {
			log.Fatal(err)
		}
		c := string(s.buffer[0])
		if c == "\n" {
			return s.out
		}
		s.out = append(s.out, s.buffer[0])
	}

	return s.out
}
