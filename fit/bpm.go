//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/tormoder/fit"
)

func main() {
	a := os.Args[1:]
	if len(a) == 0 {
		fatal(fmt.Errorf("bpm file.fit..\nwrites bpm.a and bpm.b\nktye/kyd/fit/bpm.go"))
	}
	fast, slow := make(map[byte]int), make(map[byte]int)
	for _, file := range a {
		Fit(file, fast, slow)
	}
	for b := byte(0); b < 255; b++ {
		fmt.Printf("%d %d %d %d\n", b, fast[b], slow[b], fast[b]+slow[b])
	}
}
func fatal(e error) {
	if e != nil {
		panic(e)
	}
}
func ftoa(f float64) string { return strconv.FormatFloat(f, 'v', -1, 64) }

var itoa = strconv.Itoa

func Fit(file string, fast, slow map[byte]int) {
	b, e := ioutil.ReadFile(file)
	fatal(e)

	var t *fit.File
	t, e = fit.Decode(bytes.NewReader(b))
	if e != nil && t == nil {
		return
	}

	var a *fit.ActivityFile
	a, e = t.Activity()
	if e != nil {
		return
	}

	if len(a.Sessions) == 0 {
		return
	}
	var time, dist uint32
	for _, s := range a.Sessions {
		time += s.TotalTimerTime // ms
		dist += s.TotalDistance  // 0.1m
	}

	pace := (float64(time) / 60000) / (float64(dist) / 100000)

	rec := a.Records
	for _, r := range rec {
		if r.HeartRate < 255 {
			b := byte(r.HeartRate)
			if pace < 7 {
				fast[b]++
			} else {
				slow[b]++
			}
		}
	}
}
