package geo

import "math"

type Point struct {
	lat  float64
	long float64
}

func NewPoint(lat float64, long float64) *Point {
	if math.Abs(lat) > 90 || math.Abs(long) > 180 {
		return nil
	}
	return &Point{
		lat:  lat,
		long: long,
	}
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
