package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
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

func Tour(db DB) {
	hu := func(u []uint32) (h []uint32) {
		h = make([]uint32, len(u)/2)
		j := 0
		for i := 0; i < len(u); i += 2 {
			h[j] = (u[i] & 0xffff0000) | (u[1+i] >> 16)
			j++
		}
		return h
	}
	enc := func(f File) {
		h := hu(f.WebMercator())
		m := make(map[uint32]bool)
		for _, k := range h {
			m[k] = true
		}
		var r []uint32
		for _, k := range h {
			if m[k] {
				r = append(r, k)
				m[k] = false
			}
		}
		fmt.Println(f.Start, math.Round(float64(f.Meters)), math.Round(float64(f.Seconds)), compress(r))
	}
	Each(db, func(i int, f File) { enc(f) })
}
func rle(s string) string {
	s += "ABCDEFAB" // 3..8
	var r []byte
	for i := 1; i<len(s)-8; i++ {
		c := s[i]
		j := 1
		for j<8&&s[i+j]==c {
			j++
			i++
		}
		r = append(r, c)
		if j == 2 {
			r = append(r, c)
		} else if j > 2 {
			r = append(r, "ABCDEF"[j-3])
		}
	}
	return string(r)
}
func compress(u []uint32) (r string) { /* -3 -2 -1 0 1 2 3 */
	a := "ghijklmnopqrstuvxyzGHIJKLMNOPQRSTUVXYZ!#$%^&*()_+"
	w := func(u uint32) {
		r += fmt.Sprintf("%08x", u)
	}
	b := func(dx, dy int32) {
		r += string(a[7*(dx+3)+dy+3])
	}
	xu := func(u uint32) int32 { return int32(u >> 16) }
	yu := func(u uint32) int32 { return int32(u & 0xffff) }
	if len(u) == 0 {
		return ""
	}
	w(u[0])
	for i := 1; i < len(u); i++ {
		dx, dy := xu(u[i])-xu(u[i-1]), yu(u[i])-yu(u[i-1])
		if dx < -3 || dx > 3 || dy < -3 || dy > 3 {
			w(u[i])
		} else {
			b(dx, dy)
		}
	}
	return rle(r)
}
