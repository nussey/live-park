package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/nussey/live-park/server/geo"

	"github.com/gorilla/mux"
)

const (
	serialPort = "/dev/cu.usbmodem1411"
	baudRate   = 9600
)

var printer chan string

type ParkingSpot struct {
	occupied bool
	loc      *geo.Point
}

func (p *ParkingSpot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Occupied  bool
		Latitude  float64
		Longitude float64
	}{
		Occupied:  p.occupied,
		Latitude:  p.loc.GetLat(),
		Longitude: p.loc.GetLong(),
	})
}

type ParkingLot struct {
	sync.Mutex

	// What format can I give you a geofence in?
	name    string
	spots   []*ParkingSpot
	entrace geo.Point
}

func newParkingLot(name string) *ParkingLot {
	var lot ParkingLot
	lot.name = name
	lot.spots = make([]*ParkingSpot, 0)

	return &lot
}

func main() {
	printer = make(chan string)

	mon := newSerialMonitor(serialPort, baudRate)

	lot := newParkingLot("Howey Lot")
	lot.addParkingSpot(geo.NewPoint(1, 1))
	lot.addParkingSpot(geo.NewPoint(30, 30))

	// Spin up the webserver
	router := mux.NewRouter()
	router.HandleFunc("/ReqSpot", lot.ReqSpotHandler).Methods("GET")
	go func() { log.Fatal(http.ListenAndServe(":8080", router)) }()

	go func() {
		for true {
			printer <- mon.readln()
		}
	}()

	for true {
		msg := <-printer
		fmt.Println(msg)
	}
}

func (pl *ParkingLot) ReqSpotHandler(w http.ResponseWriter, r *http.Request) {
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

	p := geo.NewPoint(lat, long)
	if p == nil {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}

	s, err := json.Marshal(pl.closestSpot(p))
	if err != nil {
		http.Error(w, "failed to marshall response", http.StatusInternalServerError)
	}

	w.Write(s)
}

func (pl *ParkingLot) closestSpot(p *geo.Point) *ParkingSpot {
	pl.Lock()
	defer pl.Unlock()

	closest := pl.spots[0]
	dis := p.DistanceTo(pl.spots[0].loc)
	for _, p2 := range pl.spots {
		if p.DistanceTo(p2.loc) < dis {
			closest = p2
		}
	}

	return closest
}

func (pl *ParkingLot) addParkingSpot(p *geo.Point) int {
	s := &ParkingSpot{
		occupied: false,
		loc:      p,
	}

	pl.Lock()
	defer pl.Unlock()

	pl.spots = append(pl.spots, s)

	return len(pl.spots) - 1
}
