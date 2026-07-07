package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

// Write shorts db for everything upto year.
// e.g. kyd -short 2018
func Shorts(db DB, year int) {
	historic := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	var shorts []uint32
	m := make(map[uint32]bool) //uniq points only
	last := int64(0)
	nold := uint64(0)
	Each(db, func(i int, f File) {
		if f.Start <= historic {
			u := f.WebMercator() //[]uint32
			for i := 0; i < len(u); i += 2 {
				k := (u[i] & 0xffff0000) | (u[1+i] >> 16)
				if m[k] == false {
					nold++
				}
				m[k] = true
			}
		}
	})
	thisyear := time.Date(1+year, time.January, 1, 0, 0, 0, 0, time.UTC).Unix()
	Each(db, func(i int, f File) {
		if f.Start > historic && f.Start < thisyear {
			u := f.WebMercator() //[]uint32
			for i := 0; i < len(u); i += 2 {
				k := (u[i] & 0xffff0000) | (u[1+i] >> 16)
				if m[k] == false {
					m[k] = true
					shorts = append(shorts, k)
				}
			}
			if f.Start > last {
				last = f.Start
			}
		}
	})
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, shorts)
	fname := fmt.Sprintf("%d.shorts", year)
	os.WriteFile(fname, b.Bytes(), 0744)
	fmt.Println("wrote", fname, "#shorts", len(shorts), "year", year, "nold", nold, "last", last)
}
