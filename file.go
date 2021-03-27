package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type File struct {
	Header
	Time []float32 // seconds
	Dist []float32 // meters
	Alt  []float32 // altitude (m)
	Lat  []int32   // semicircles (invalid: 0x7FFFFFFF) (180 / math.Pow(2, 31))
	Lon  []int32   // semicircles
}
type Header struct {
	Start   int64   // unix time (seconds)
	Type    uint32  // type 1(run) 2(cycle)
	Seconds float32 // total duration
	Meters  float32 // total distance
	Samples uint64  // number of samples
}
type Race struct {
	Start   int64   // unix time (seconds)
	Type    string  // "800m"
	Meters  float32 // meters
	Seconds float32 // seconds
	Result  string  // "101/2048"
	Name    string
}

func (f *File) alloc() {
	samples := f.Samples
	f.Time = make([]float32, samples)
	f.Dist = make([]float32, samples)
	f.Alt = make([]float32, samples)
	f.Lat = make([]int32, samples)
	f.Lon = make([]int32, samples)
}
func (f File) Empty() bool { return f.Start == 0 }
func (f File) WebMercator() (lat []uint32, lon []uint32) {
	lat = make([]uint32, 0, f.Samples)
	lon = make([]uint32, 0, f.Samples)
	for i := uint64(0); i < f.Samples; i++ {
		if a, b := mercator(f.Lat[i], f.Lon[i]); math.IsNaN(a) == false {
			lat = append(lat, uint32(a))
			lon = append(lon, uint32(b))
		}
	}
	return lat, lon
}
func mercator(lat, lon int32) (float64, float64) {
	const minLat float64 = -85.0511287725758 // (zoom=24, -84.92832092949963 zoom=0)
	const maxLat float64 = 85.05112877980659
	if lat == invalidSemis || lon == invalidSemis {
		return math.NaN(), math.NaN()
	}
	la, lo := Deg(lat), Deg(lon)
	if la < minLat || la > maxLat {
		return math.NaN(), math.NaN()
	}
	x := (lo + 180) / 360
	y := (1 - math.Log(math.Tan(rad(la)+1/math.Cos(rad(la))))/math.Pi) / 2
	return x, y
}
func rad(deg float64) float64 { return math.Pi * deg / 180.0 }

func Deg(s int32) float64 {
	if s == invalidSemis {
		return math.NaN()
	}
	return float64(s) * 8.381903171539307e-08 // * 180/2^31
}
func ParseHeader(s string) (h Header, e error) {
	v := strings.Fields(s)
	err := func(s string) error { return fmt.Errorf("index: %s", s) }
	if len(v) != 5 {
		return h, err(fmt.Sprintf("expected %d fields (not %d)", 5, len(v)))
	}
	h.Start, e = strconv.ParseInt(v[0], 10, 64)
	if e != nil {
		return h, err("parse start")
	}
	var u uint64
	u, e = strconv.ParseUint(v[1], 10, 32)
	h.Type = uint32(u)
	if e != nil {
		return h, err("parse type")
	}
	var f float64
	f, e = strconv.ParseFloat(v[2], 32)
	h.Seconds = float32(f)
	if e != nil {
		return h, err("parse seconds")
	}
	f, e = strconv.ParseFloat(v[3], 32)
	h.Meters = float32(f)
	if e != nil {
		return h, err("parse meters")
	}
	h.Samples, e = strconv.ParseUint(v[4], 10, 64)
	if e != nil {
		return h, err("samples")
	}
	return h, nil
}
func (h Header) Indexline() string { // entry(line) in db/index.txt
	return fmt.Sprint(h.Start, h.Type, h.Seconds, h.Meters, h.Samples)
}
func (h Header) String() string { // list output
	date := time.Unix(h.Start, 0).Format("2006.01.02T15:04:05")
	hh := int(h.Seconds / 3600)
	mm := int(h.Seconds/60) - hh*60
	ss := int(h.Seconds) - hh*3600 - mm*60
	return fmt.Sprintf("%d %c %s %02d:%02d:%02d %v", h.Start, sport(h.Type), date, hh, mm, ss, h.Meters/1000)
}
func Decode(b []byte) (File, error) {
	r := bytes.NewReader(b)
	var f File
	if e := binary.Read(r, le, &f.Header); e != nil {
		return f, e
	}
	f.alloc()
	var e error
	e = do(e, binary.Read(r, le, f.Time))
	e = do(e, binary.Read(r, le, f.Dist))
	e = do(e, binary.Read(r, le, f.Alt))
	e = do(e, binary.Read(r, le, f.Lat))
	e = do(e, binary.Read(r, le, f.Lon))
	return f, e
}
func (f File) Encode(w io.Writer) (e error) {
	e = do(e, binary.Write(w, le, f.Header))
	e = do(e, binary.Write(w, le, f.Time))
	e = do(e, binary.Write(w, le, f.Dist))
	e = do(e, binary.Write(w, le, f.Alt))
	e = do(e, binary.Write(w, le, f.Lat))
	e = do(e, binary.Write(w, le, f.Lon))
	return e
}
func do(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

func (f File) Table(w io.Writer) {
	fmt.Fprintf(w, "Start:   %s (%d)\n", time.Unix(f.Start, 0).Format("2006.01.02T15:04:05"), f.Start)
	fmt.Fprintf(w, "Type:    %c\n", sport(f.Type))
	fmt.Fprintf(w, "Seconds: %v (%s)\n", f.Seconds, time.Duration(f.Seconds)*time.Second)
	fmt.Fprintf(w, "Meters:  %v\n", f.Meters)
	tw := tabwriter.NewWriter(w, 2, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "#\tTime\tDist\tAlt\tLat\tLon\n")
	for i := 0; i < int(f.Samples); i++ {
		fmt.Fprintf(tw, "%d\t%v\t%v\t%v\t%.6f\t%.6f\n", i, f.Time[i], f.Dist[i], f.Alt[i], Deg(f.Lat[i]), Deg(f.Lon[i]))
	}
	tw.Flush()
}
func sport(x uint32) (r byte) {
	r = '?'
	switch x {
	case 1:
		r = 'R'
	case 2:
		r = 'B'
	}
	return r
}

const (
	invalidSemis int32 = 0x7FFFFFFF
)

var le = binary.LittleEndian
