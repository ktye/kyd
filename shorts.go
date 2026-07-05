package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// Write initial shorts db with kyd -short 1 (since timestamp 1)
// and updates with the id from the last update output name: kyd -short 1783100117
func Shorts(db DB, since int64) {
	var shorts []uint32
	m := make(map[uint32]bool) //uniq points only
	last := int64(0)
	nold := uint64(0)
	Each(db, func(i int, f File) {
		if f.Start <= since {
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
	Each(db, func(i int, f File) {
		if f.Start > since {
			u := f.WebMercator() //[]uint32
			for i := 0; i < len(u); i += 2 {
				k := (u[i] & 0xffff0000) | (u[1+i] >> 16)
				if m[k] == false {
					m[k] = true
					shorts = append(shorts, k)
				}
			}
		}
		if f.Start > last {
			last = f.Start
		}

	})
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, shorts)
	fname := fmt.Sprintf("%d.shorts", last)
	os.WriteFile(fname, b.Bytes(), 0744)
	fmt.Println("wrote", fname, "#shorts", len(shorts), "since", since, "nold", nold, "last", last)
}
