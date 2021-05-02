package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/tormoder/fit"
)

func ReadFit(file string) (f File, e error) {
	var b []byte
	b, e = ioutil.ReadFile(file)
	if e != nil {
		return
	}

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
	if len(rec) < 1 {
		return f, fmt.Errorf("file has no records")
	}
	ses := a.Sessions
	var start time.Time
	var sport uint32
	if len(ses) > 0 {
		start = ses[0].StartTime
		sport = uint32(ses[0].Sport)
	}

	var seconds, meters uint32
	for _, s := range ses {
		seconds += s.TotalTimerTime
		meters += s.TotalDistance
	}

	samples := len(rec)
	if start.IsZero() && samples > 0 {
		start = rec[0].Timestamp
	}
	f.Header = Header{
		Start:   start.Unix(),
		Type:    sport,
		Seconds: float32(seconds) / 1000.0,
		Meters:  float32(meters) / 100.0,
		Samples: uint64(samples),
	}
	f.alloc()

	for i, r := range rec {
		f.Time[i] = float32(r.Timestamp.Sub(start).Seconds())
		f.Dist[i] = float32(r.GetDistanceScaled())
		f.Alt[i] = float32(r.GetEnhancedAltitudeScaled())
		f.Lat[i] = r.PositionLat.Semicircles()
		f.Lon[i] = r.PositionLong.Semicircles()
	}
	return f, nil
}
