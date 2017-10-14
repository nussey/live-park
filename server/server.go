package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/nussey/live-park/server/geo"

	"github.com/gorilla/mux"
)

const (
	serialPort = "/dev/cu.usbmodem1411"
	baudRate   = 9600
)

var printer chan string

func main() {
	printer = make(chan string)

	mon := newSerialMonitor(serialPort, baudRate)

	lot, err := newParkingLot("./HoweyLot.json")
	if err != nil {
		log.Fatal("Failed to read JSON: ", err)
	}

	// Spin up the webserver
	router := mux.NewRouter()
	router.HandleFunc("/ReqSpot", lot.ReqSpotHandler).Methods("GET")
	router.HandleFunc("/LotInfo", lot.LotDataHandler).Methods("GET")
	go func() { log.Fatal(http.ListenAndServe(":8080", router)) }()

	go func() {
		for true {
			printer <- mon.readln()
		}
	}()

	// Print stuff out
	for true {
		msg := <-printer
		fmt.Println(msg)
	}
}

func newParkingLot(filename string) (*ParkingLot, error) {
	var lot ParkingLot

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(file, &lot)
	if err != nil {
		return nil, err
	}

	return &lot, nil
}

func (pl *ParkingLot) LotDataHandler(w http.ResponseWriter, r *http.Request) {
	s, err := json.Marshal(pl)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Write(s)
}

func (pl *ParkingLot) ReqSpotHandler(w http.ResponseWriter, r *http.Request) {
	// Parse input
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	if err != nil {
		http.Error(w, "Unable to parse latitude", http.StatusBadRequest)
		return
	}
	long, err := strconv.ParseFloat(r.URL.Query().Get("long"), 64)
	if err != nil {
		http.Error(w, "Unable to parse longitude", http.StatusBadRequest)
		return
	}

	// Create a point to represent the car
	p := geo.NewPoint(lat, long)
	if p == nil {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}

	s, err := json.Marshal(pl.closestSpot(p).Location)
	if err != nil {
		http.Error(w, "failed to marshall response", http.StatusInternalServerError)
	}

	w.Write(s)
}

func (pl *ParkingLot) closestSpot(p *geo.Point) *ParkingSpot {
	pl.Lock()
	defer pl.Unlock()

	closest := pl.spots[0]
	dis := p.DistanceTo(&pl.spots[0].Location)
	for _, p2 := range pl.spots {
		if p.DistanceTo(&p2.Location) < dis {
			closest = p2
		}
	}

	return closest
}

func (pl *ParkingLot) addParkingSpot(id int64, p *geo.Point) int {
	s := &ParkingSpot{
		occupied:   false,
		reserved:   false,
		Location:   *p,
		HardwareId: id,
	}

	pl.Lock()
	defer pl.Unlock()

	pl.spots[id] = s

	return len(pl.spots) - 1
}
