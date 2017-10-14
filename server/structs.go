package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/nussey/live-park/server/geo"
)

type ParkingSpot struct {
	occupied bool
	reserved bool

	Location geo.Point
	Name     string

	HardwareId int32
	battery    int
}

type ParkingLot struct {
	sync.Mutex

	name string
	// dollars per hour
	price float32

	spots map[int32]*ParkingSpot

	entrance geo.Point
	geofence []geo.Point
}

func spotsToMap(spots []*ParkingSpot) map[int32]*ParkingSpot {
	m := make(map[int32]*ParkingSpot)
	for _, spot := range spots {
		m[spot.HardwareId] = spot
	}

	return m
}

func mapToSpots(m map[int32]*ParkingSpot) []*ParkingSpot {
	spots := make([]*ParkingSpot, len(m))

	i := 0
	for _, spot := range m {
		spots[i] = spot
		i++
	}

	return spots
}

func (p *ParkingLot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name           string
		Price          float32
		AvailableSpots int
		TotalSpots     int

		Entrance geo.Point
		GeoFence []geo.Point
	}{
		Name:           p.name,
		Price:          p.price,
		AvailableSpots: p.AvailableSpots(),
		TotalSpots:     p.TotalSpots(),

		Entrance: p.entrance,
		GeoFence: p.geofence,
	})
}

func (p *ParkingLot) UnmarshalJSON(b []byte) error {
	type unmar struct {
		Name     string
		Price    float32
		Entrance geo.Point
		GeoFence []geo.Point
		Spots    []*ParkingSpot
	}

	var d unmar

	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}

	p.name = d.Name
	p.price = d.Price
	p.entrance = d.Entrance
	p.geofence = d.GeoFence
	p.spots = spotsToMap(d.Spots)

	return nil
}

func (p *ParkingSpot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name      string
		Latitude  float64
		Longitude float64
		Available bool
	}{
		Name:      p.Name,
		Latitude:  p.Location.GetLat(),
		Longitude: p.Location.GetLong(),
		Available: !(p.occupied || p.reserved),
	})
}

func (p *ParkingSpot) UnmarshalJSON(b []byte) error {
	type unmar struct {
		Name        string
		HardwareId  int32
		Coordinates string
	}

	var d unmar

	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}

	p.Name = d.Name
	p.HardwareId = d.HardwareId
	coordinates := strings.Split(d.Coordinates, ", ")
	lat, err := strconv.ParseFloat(coordinates[0], 64)
	if err != nil {
		return err
	}

	long, err := strconv.ParseFloat(coordinates[1], 64)
	if err != nil {
		return err
	}
	p.Location = *geo.NewPoint(lat, long)

	return nil
}

type SerialPayload struct {
	Identifier        int32
	Occupied          int
	BatteryPercentage int
}
