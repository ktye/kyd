package main

import (
	"fmt"
)

func K(db DB) {
	// id type year week day dist time coords
	// i  s    i    i    i   f    f    i
	n := db.Len()
	for i := 0; i<n; i++ {
		h := db.Head(i)
		id := h.Start // int32 til 2038
		typ := sport(h.Type) + 32
		yw := h.yearweek()
		year := yw.Year
		week := yw.Week - 1
		day := h.day()
		time := h.Meters / 1000.0
		dist := h.Seconds / 3600.0
		coords := h.Samples
		fmt.Printf("%d %c %d %d %d %.3f %.4f %d\n", id, typ, year, week, day, dist, time, coords)
	}
}
