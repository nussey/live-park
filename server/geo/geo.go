package geo

import (
	"encoding/json"
	"errors"
	"math"
)

type Point struct {
	lat  float64
	long float64
}

func (p *Point) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Latitude  float64
		Longitude float64
	}{
		Latitude:  p.lat,
		Longitude: p.long,
	})
}

func (p *Point) UnmarshalJSON(b []byte) error {
	unmar := &struct {
		Latitude  float64
		Longitude float64
	}{}
	err := json.Unmarshal(b, unmar)
	if err != nil {
		return err
	}

	p.lat = unmar.Latitude
	p.long = unmar.Longitude
	if !p.checkPoint() {
		return errors.New("Invalid point")
	}
	return nil
}

func NewPoint(lat float64, long float64) *Point {
	p := &Point{
		lat:  lat,
		long: long,
	}
	if !p.checkPoint() {
		return nil
	}
	return p
}

func (p *Point) checkPoint() bool {
	if math.Abs(p.lat) > 90 || math.Abs(p.long) > 180 {
		return false
	}
	return true
}

func (p *Point) GetLat() float64 {
	return p.lat
}

func (p *Point) GetLong() float64 {
	return p.long
}

func (p *Point) DistanceTo(p2 *Point) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = p.lat * math.Pi / 180
	lo1 = p.long * math.Pi / 180
	la2 = p2.lat * math.Pi / 180
	lo2 = p2.long * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
