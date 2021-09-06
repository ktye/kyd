package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/tormoder/fit"
)

type recfunc func(fit.RecordMsg) string

func (f recfunc) String(r fit.RecordMsg) string { return f(r) }

func main() {
	a := os.Args[1:]
	f, h, t, r := recprint, "", "", false
	if len(a) > 1 && a[0] == "gpx" {
		a = a[1:]
		f = gpxprint
		h = `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.0"><trk><trkseg>` + "\n"
		t = `</trkseg></trk></gpx>` + "\n"
	}
	if len(a) > 1 && a[0] == "3x" {
		a = a[1:]
		r = true
	}
	if len(a) < 1 {
		fatal(fmt.Errorf("fit file.fit..\nfit gpx file.fit..\nktye/kyd/fit/main.go"))
	}
	for _, file := range a {
		Fit(file, f, h, t, r)
	}
}
func recprint(r *fit.RecordMsg) {
	// https://pkg.go.dev/github.com/tormoder/fit#RecordMsg
	fmt.Println(r.Timestamp.Format("2006.01.02T15:04:05"), r.GetDistanceScaled()/1000.0, r.PositionLat.Degrees(), r.PositionLong.Degrees(), r.GetEnhancedAltitudeScaled(), r.HeartRate)
}
func gpxprint(r *fit.RecordMsg) {
	fmt.Printf("<trkpt lat=\"%.8f\" lon=\"%.8f\"><time>%s</time></trkpt>\n", r.PositionLat.Degrees(), r.PositionLong.Degrees(), r.Timestamp.Format("2006-01-02T15:04:05Z"))
}
func fatal(e error) {
	if e != nil {
		panic(e)
	}
}
func ftoa(f float64) string { return strconv.FormatFloat(f, 'v', -1, 64) }

var itoa = strconv.Itoa

func Fit(file string, recprint func(r *fit.RecordMsg), head, tail string, reduce bool) {
	b, e := ioutil.ReadFile(file)
	fatal(e)

	var t *fit.File
	t, e = fit.Decode(bytes.NewReader(b))
	if e != nil {
		fmt.Println(e) // use partial file
		if t == nil {
			return
		}
	}

	var a *fit.ActivityFile
	a, e = t.Activity()
	if e != nil {
		return
	}

	rec := a.Records
	os.Stdout.Write([]byte(head))
	rr := 0
	for _, r := range rec {
		if !reduce || rr == 2 {
			recprint(r)
			rr = -1
		}
		rr++
	}
	os.Stdout.Write([]byte(tail))
}
