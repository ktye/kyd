package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

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
		return
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
	if len(ses) < 1 {
		return f, fmt.Errorf("file has no sessions")
	}

	var seconds, meters uint32
	for _, s := range ses {
		seconds += s.TotalTimerTime
		meters += s.TotalDistance
	}

	samples := len(rec)
	start := ses[0].StartTime
	f.Header = Header{
		Start:   start.Unix(),
		Type:    uint32(ses[0].Sport),
		Seconds: float32(seconds) / 1000.0,
		Meters:  float32(meters) / 100.0,
		Samples: uint64(samples),
	}
	f.alloc()

	for i, r := range rec {
		f.Time[i] = float32(r.Timestamp.Sub(start).Seconds())
		f.Dist[i] = float32(r.GetDistanceScaled())
		f.Alt[i] = float32(r.GetAltitudeScaled())
		f.Lat[i] = r.PositionLat.Semicircles()
		f.Lon[i] = r.PositionLong.Semicircles()
	}
	return f, nil
}
