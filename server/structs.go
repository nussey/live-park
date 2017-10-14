package main

import (
	"encoding/json"
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

	entrace  geo.Point
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

		Entrace  geo.Point
		GeoFence []geo.Point
		Spots    []*ParkingSpot
	}{
		Name:           p.name,
		Price:          p.price,
		AvailableSpots: p.AvailableSpots(),
		TotalSpots:     p.TotalSpots(),

		Entrace:  p.entrace,
		GeoFence: p.geofence,
		Spots:    mapToSpots(p.spots),
	})
}

func (p *ParkingSpot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name      string
		Latitude  float64
		Longitude float64
	}{
		Name:      p.Name,
		Latitude:  p.Location.GetLat(),
		Longitude: p.Location.GetLong(),
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
	p.entrace = d.Entrance
	p.geofence = d.GeoFence
	p.spots = spotsToMap(d.Spots)

	return nil
}

type SerialPayload struct {
	Identifier        int32
	Occupied          int
	BatteryPercentage int
}
