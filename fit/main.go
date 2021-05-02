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
	if len(os.Args) < 2 {
		fatal(fmt.Errorf("fit file.fit..\nktye/kyd/fit/main.go"))
	}
	for _, file := range os.Args[1:] {
		Fit(file, recprint)
	}
}
func recprint(r *fit.RecordMsg) {
	// https://pkg.go.dev/github.com/tormoder/fit#RecordMsg
	fmt.Println(r.Timestamp.Format("2006.01.02T15:04:05"), r.GetDistanceScaled()/1000.0, r.PositionLat.Degrees(), r.PositionLong.Degrees(), r.GetEnhancedAltitudeScaled(), r.HeartRate)
}
func fatal(e error) {
	if e != nil {
		panic(e)
	}
}
func ftoa(f float64) string { return strconv.FormatFloat(f, 'v', -1, 64) }

var itoa = strconv.Itoa

func Fit(file string, recprint func(r *fit.RecordMsg)) {
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

	for _, r := range rec {
		recprint(r)
	}
}
