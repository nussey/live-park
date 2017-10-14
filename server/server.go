package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	reserved bool
	loc      geo.Point

	id      int64
	battery int
}

type ParkingLot struct {
	sync.Mutex

	name string
	// dollars per hour
	price float32

	spots []*ParkingSpot

	entrace  geo.Point
	geofence []geo.Point
}

func (p *ParkingLot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name           string
		Price          float32
		AvailableSpots int
		TotalSpots     int

		Entrace  geo.Point
		GeoFence []geo.Point
	}{
		Name:           p.name,
		Price:          p.price,
		AvailableSpots: 0,
		TotalSpots:     1,

		Entrace:  p.entrace,
		GeoFence: p.geofence,
	})
}

func (p *ParkingLot) UnmarshalJSON(b []byte) error {
	type unmar struct {
		Name     string
		Price    float32
		Entrance geo.Point
		GeoFence []geo.Point
		Spots    []geo.Point
	}

	var d unmar

	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}

	p.name = d.Name
	p.price = d.Price
	p.entrace = d.Entrance
	p.geofence = d.GeoFence
	for _, spot := range d.Spots {
		p.addParkingSpot(&spot)
	}

	return nil
}

// func newParkingLot(name string, entrace *geo.Point, fence []geo.Point) *ParkingLot {
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

	// lot.name = name
	// lot.spots = make([]*ParkingSpot, 0)
	// lot.entrace = *entrace
	// lot.geofence = fence

	return &lot, nil
}

func main() {
	printer = make(chan string)

	mon := newSerialMonitor(serialPort, baudRate)

	// fence := []geo.Point{*geo.NewPoint(0, 0),
	// 	*geo.NewPoint(50, 0),
	// 	*geo.NewPoint(50, 50),
	// 	*geo.NewPoint(0, 50)}
	// lot := newParkingLot("Howey Lot", geo.NewPoint(10, 10), fence)
	lot, err := newParkingLot("./HoweyLot.json")
	if err != nil {
		log.Fatal("Failed to read JSON: ", err)
	}
	// lot.addParkingSpot(geo.NewPoint(1, 1))
	// lot.addParkingSpot(geo.NewPoint(30, 30))

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

	s, err := json.Marshal(pl.closestSpot(p).loc)
	if err != nil {
		http.Error(w, "failed to marshall response", http.StatusInternalServerError)
	}

	w.Write(s)
}

func (pl *ParkingLot) closestSpot(p *geo.Point) *ParkingSpot {
	pl.Lock()
	defer pl.Unlock()

	closest := pl.spots[0]
	dis := p.DistanceTo(&pl.spots[0].loc)
	for _, p2 := range pl.spots {
		if p.DistanceTo(&p2.loc) < dis {
			closest = p2
		}
	}

	return closest
}

func (pl *ParkingLot) addParkingSpot(p *geo.Point) int {
	s := &ParkingSpot{
		occupied: false,
		reserved: false,
		loc:      *p,
	}

	pl.Lock()
	defer pl.Unlock()

	pl.spots = append(pl.spots, s)

	return len(pl.spots) - 1
}
