package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

const (
	serialPort = "/dev/cu.usbmodem14221"
	baudRate   = 9600
)

var printer chan string

type ParkingSpot struct {
	occupied bool
}

type ParkingLot struct {
	sync.Mutex

	// What format can I give you a geofence in?
	name  string
	spots []ParkingSpot
}

func newParkingLot(name string, spots int) *ParkingLot {
	var lot ParkingLot
	lot.name = name
	lot.spots = make([]ParkingSpot, spots)

	return &lot
}

func main() {
	printer = make(chan string)

	mon := newSerialMonitor(serialPort, baudRate)

	lot := newParkingLot("Howey Lot", 10)

	router := mux.NewRouter()
	router.HandleFunc("/foobar", lot.TestHandler).Methods("GET")
	go func() { log.Fatal(http.ListenAndServe(":8080", router)) }()

	go func() {
		for true {
			printer <- "diddle"
			printer <- mon.readln()
		}
	}()

	for true {
		msg := <-printer
		fmt.Println(msg)
	}
}

func (pl *ParkingLot) TestHandler(w http.ResponseWriter, r *http.Request) {
	printer <- "FOOBAR"
}
