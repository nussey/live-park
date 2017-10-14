package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/nussey/live-park/server/geo"

	"github.com/gorilla/mux"
)

const (
	serialPort = "/dev/cu.usbmodem1421"
	baudRate   = 9600
)

var printer chan string

func main() {
	printer = make(chan string)

	lot, err := newParkingLot("./HoweyLot.json")
	if err != nil {
		log.Fatal("Failed to read JSON: ", err)
	}

	// Spin up the webserver
	router := mux.NewRouter()
	router.HandleFunc("/ReqSpot", lot.ReqSpotHandler).Methods("GET")
	router.HandleFunc("/LotInfo", lot.LotDataHandler).Methods("GET")
	router.HandleFunc("/SpotList", lot.SpotListHandler).Methods("GET")
	go func() { log.Fatal(http.ListenAndServe(":8080", router)) }()

	go monitorSerial(lot)

	// Print stuff out
	for true {
		msg := <-printer
		fmt.Println(msg)
	}
}

func monitorSerial(lot *ParkingLot) {
	mon := newSerialMonitor(serialPort, baudRate)
	for true {
		str := mon.readln()
		var payload SerialPayload
		err := json.Unmarshal(str, payload)
		if err != nil {
			printer <- "ERROR UNMARSHALING FROM SERIAL"
			continue
		}

		spot := lot.GetSpot(payload.Identifier)
		if spot == nil {
			printer <- fmt.Sprintf("Invalid hardware ID %d", payload.Identifier)
			continue
		}
		spot.battery = payload.BatteryPercentage
		state := false
		if payload.Occupied > 0 {
			state = true
		}

		if state != spot.occupied {
			var err error
			if state {
				err = lot.TakeSpot(spot.HardwareId)
			} else {
				err = lot.LeaveSpot(spot.HardwareId)
			}
			if err != nil {
				printer <- ("ERROR CHANGING SPOT STATE: " + err.Error())
			}
		}

		printer <- fmt.Sprintf("%d: %t", payload.Identifier, payload.Occupied)
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

func (pl *ParkingLot) GetSpot(id int32) *ParkingSpot {
	pl.Lock()
	defer pl.Unlock()

	return pl.spots[id]
}

// Mark that a car pulled in to a spot
// Threadsafe
func (pl *ParkingLot) TakeSpot(id int32) error {
	pl.Lock()
	defer pl.Unlock()

	spot := pl.spots[id]

	// if spot.occupied {
	// 	return errors.New("Spot already occupied")
	// }

	spot.reserved = false
	spot.occupied = true

	return nil
}

// Mark that a spot is reserved for a car coming in
// Threadsafe
func (pl *ParkingLot) ReserveSpot(id int32) error {
	pl.Lock()
	defer pl.Unlock()

	spot := pl.spots[id]

	if spot.reserved || spot.occupied {
		return errors.New("Spot already reserved or occupied")
	}

	spot.reserved = true
	spot.occupied = false

	return nil
}

// Mark that a car just left a spot
// Threadsafe
func (pl *ParkingLot) LeaveSpot(id int32) error {
	pl.Lock()
	defer pl.Unlock()

	spot := pl.spots[id]

	// if !spot.occupied {
	// 	return errors.New("Spot isn't taken!")
	// }

	spot.reserved = false
	spot.occupied = false

	return nil
}

// Count the number of empty spots in the lot
// Threadsafe
func (pl *ParkingLot) AvailableSpots() int {
	pl.Lock()
	defer pl.Unlock()
	empty := 0
	for _, spot := range pl.spots {
		if !spot.occupied && !spot.reserved {
			empty++
		}
	}

	return empty
}

// Count the number of spots in the lot
// Threadsafe
func (pl *ParkingLot) TotalSpots() int {
	pl.Lock()
	defer pl.Unlock()
	return len(pl.spots)
}

func (pl *ParkingLot) LotDataHandler(w http.ResponseWriter, r *http.Request) {
	s, err := json.Marshal(pl)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Write(s)
}

func (pl *ParkingLot) SpotListHandler(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(mapToSpots(pl.spots))
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
	}

	w.Write(b)
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

	if pl.AvailableSpots() == 0 {
		http.Error(w, "No free spots remaining!", http.StatusBadRequest)
		return
	}

	// Reserve the closest spot
	spot := pl.closestSpot(p)
	pl.ReserveSpot(spot.HardwareId)

	s, err := json.Marshal(spot)
	if err != nil {
		http.Error(w, "failed to marshall response", http.StatusInternalServerError)
		return
	}

	w.Write(s)
}

func (pl *ParkingLot) closestSpot(p *geo.Point) *ParkingSpot {
	pl.Lock()
	defer pl.Unlock()

	var closest *ParkingSpot
	dis := 2147483647.0
	for _, spot := range pl.spots {
		if !spot.occupied && !spot.reserved && p.DistanceTo(&spot.Location) < dis {
			dis = p.DistanceTo(&spot.Location)
			closest = spot
		}
	}

	return closest
}

func (pl *ParkingLot) addParkingSpot(id int32, p *geo.Point) int {
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
