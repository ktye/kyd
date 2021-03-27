package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	near        = 50.0 // 50m
	EarthRadius = 6371008.8
)

func Here(db DB, s string) SubDB {
	v := strings.Split(s, ",")
	if len(v) != 2 {
		fatal(fmt.Errorf("here: expect lat,lon (got %s)", s))
	}
	la, lo := rad(parseFloat(v[0])), rad(parseFloat(v[1]))

	r := Filter(db, func(f File) bool {
		for i := uint64(0); i < f.Samples; i++ {
			lat, lon := rad(Deg(f.Lat[i])), rad(Deg(f.Lon[i]))
			if !math.IsNaN(lat) && !math.IsNaN(lon) {
				m := Vincenty(la, lo, lat, lon)
				if math.IsNaN(m) == false && m < near {
					return true
				}
			}
		}
		return false
	})
	return r
}

func Vincenty(lat1, lon1, lat2, lon2 float64) (meters float64) {
	dLon := math.Abs(lon2 - lon1)
	a := math.Cos(lat2) * math.Sin(dLon)
	b := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	return EarthRadius * math.Atan2(math.Sqrt(a*a+b*b), math.Sin(lat1)*math.Sin(lat2)+math.Cos(lat1)*math.Cos(lat2)*math.Cos(dLon))
}

func parseFloat(s string) float64 {
	f, e := strconv.ParseFloat(s, 64)
	fatal(e)
	return f
}
