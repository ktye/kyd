package main

import (
	"fmt"
)

func K(db DB) {
	// id type year week day dist time coords
	// i  s    i    i    i   f    f    i
	fmt.Println("id,type,time,dist,coords")
	fmt.Println("i,s,f,f,i")
	n := db.Len()
	for i := 0; i < n; i++ {
		h := db.Head(i)
		id := h.Start // int32 til 2038
		typ := sport(h.Type) + 32
		time := h.Seconds / 3600.0
		dist := h.Meters / 1000.0
		coords := h.Samples
		fmt.Printf("%d,%c,%.3f,%.4f,%d\n", id, typ, time, dist, coords)
	}
}
