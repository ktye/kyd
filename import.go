package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func importDB(src, dst string) {
	if d, e := ioutil.ReadDir(dst); e != nil {
		fatal(e)
	} else if len(d) != 2 {
		fatal(fmt.Errorf("import: dst is not empty: %s (should contain an empty index/race.txt)", dst))
	}
	if d, e := ioutil.ReadFile(filepath.Join(dst, "index.txt")); len(d) != 0 || e != nil {
		fatal(fmt.Errorf("import: dst must contain an empty index.txt #%d e=%v", len(d), e))
	}

	heads, race := importIndex(src)
	fmt.Println(len(heads), len(race))

	db, e := OpenDB(dst)
	fatal(e)

	for _, h := range heads {
		if f, e := importJson(src, h); e == nil || os.IsNotExist(e) {
			db.Add(f)
			if (f.Samples == 0) != os.IsNotExist(e) {
				fatal(fmt.Errorf("%s: %d sample, err = %v\n", f.Header.String(), f.Samples, e))
			}
			if f.Samples > 0 {
				fmt.Println(f.Start, f.Samples)
			}
		} else {
			fatal(e)
		}
	}

	f, e := os.Create(db.racepath())
	fatal(e)
	defer f.Close()
	for _, r := range race {
		fmt.Fprintln(f, r.String())
	}
}
func importJson(dir string, h Header) (f File, err error) {
	f.Header = h
	b, e := ioutil.ReadFile(filepath.Join(dir, unix(h.Start).Format("2006/20060102T150405.json")))
	if e != nil {
		return f, e
	}
	keys := []string{"start", "type", "title", "time", "dist", "desc", "lap", "track", "points", "lat", "lon", "elev"}
	for _, k := range keys {
		b = bytes.Replace(b, []byte(k), []byte(`"`+k+`"`), -1)
	}
	b = bytes.Replace(b, []byte{9}, nil, -1)
	b = bytes.Replace(b, []byte{10}, nil, -1)
	b = bytes.Replace(b, []byte{32}, nil, -1)
	b = bytes.Replace(b, []byte(",}"), []byte("}"), -1)
	b = bytes.Replace(b, []byte(",]"), []byte("]"), -1)
	if len(b) > 0 && b[len(b)-1] == ',' {
		b = b[:len(b)-1]
	}
	b = bytes.Replace(b, []byte("undefined"), []byte("-123456"), -1)
	b = append([]byte{'{'}, b...)
	b = append(b, '}')

	type tk struct {
		Points int      `json:points`
		Time   []jfloat `json:time`
		Dist   []jfloat `json:dist`
		Lat    []jfloat `json:lat`
		Lon    []jfloat `json:lon`
		Elev   []jfloat `json:elev`
	}
	type l struct {
		Start int    `json:start`
		Time  jfloat `json:time`
		Dist  jfloat `json:dist`
		Track tk     `json:track`
	}
	type t struct {
		Start string `json:start`
		Type  string `json:type`
		Time  string `json:time`
		Dist  jfloat `json:dist`
		Lap   []l    `json:lap`
	}
	var d t
	if e := json.Unmarshal(b, &d); e != nil {
		fmt.Println(string(b))
		fatal(e)
	}

	for _, l := range d.Lap {
		tk := l.Track
		f.Time = append(f.Time, jfloats32(tk.Time)...)
		f.Dist = append(f.Dist, jfloats32(tk.Dist)...)
		f.Lat = append(f.Lat, jsemis(tk.Lat)...)
		f.Lon = append(f.Lat, jsemis(tk.Lon)...)
	}

	samples := len(f.Time)
	if len(f.Dist) != samples || len(f.Alt) != samples || len(f.Lat) != samples || len(f.Lon) != samples {
		return f, fmt.Errorf("uniform")
	}

	f.Samples = uint64(samples)
	return f, nil
}
func importIndex(dir string) (h []Header, r []Race) {
	index, e := ioutil.ReadFile(filepath.Join(dir, "index.json"))
	fatal(e)
	keys := []string{"type", "time", "dist", "climb", "laps", "agresult", "result", "racetime", "racetype", "title", "list", "links"}
	for _, k := range keys {
		index = bytes.Replace(index, []byte(","+k), []byte(`,"`+k+`"`), -1)
	}
	index = bytes.Replace(index, []byte("file"), []byte(`"file"`), -1)
	index = bytes.Replace(index, []byte(",}"), []byte("}"), -1)
	index = append([]byte{'['}, index...)
	index = index[:len(index)-2] // trailing ,
	index = append(index, ']')

	racetime := func(s string) time.Duration {
		if s == "DNF" || s == "" {
			return 0
		}
		if len(s) != 8 {
			fmt.Printf("time? %q\n", s)
			panic("parse time")
		}
		b := []byte(s)
		b[2] = 'h'
		b[5] = 'm'
		s = string(b) + "s"
		d, e := time.ParseDuration(s)
		fatal(e)
		return d
	}
	fileTime := func(s string) int64 {
		s = strings.TrimSuffix(s[5:], ".json")
		if s == "20150631T193000" {
			s = "20150701T193000"
		}
		t, e := time.Parse("20060102T150405", s)
		fatal(e)
		return t.Unix()
	}
	parseType := func(s string) uint32 {
		switch s {
		case "R":
			return 1
		case "B":
			return 2
		case "S":
			return 5
		}
		panic("unknown type: " + s)
	}
	parseFloat32 := func(s string) float32 {
		f, e := strconv.ParseFloat(s, 32)
		fatal(e)
		return float32(f)
	}
	type hdr struct {
		File     string `json:file`
		Type     string `json:type`
		Time     string `json:time`
		Dist     string `json:dist`
		Result   string `json:result`
		Racetime string `json:racetime`
		Racetype string `json:racetype`
		Title    string `json:title`
	}
	race := func(x hdr) Race {
		return Race{
			Start:  fileTime(x.File),
			Type:   x.Racetype,
			Time:   racetime(x.Racetime),
			Result: x.Result,
			Name:   x.Title,
		}
	}
	head := func(x hdr) Header {
		if x.Dist == "" {
			fmt.Println(x)
		}
		return Header{
			Start:   fileTime(x.File),
			Type:    parseType(x.Type),
			Seconds: float32(racetime(x.Time).Seconds()),
			Meters:  parseFloat32(x.Dist),
		}
	}
	var d []hdr
	fatal(json.Unmarshal(index, &d))

	for _, x := range d {
		if x.Type == "C" {
			r = append(r, race(x))
		} else {
			if x.File == "2015/20150823T171833.json" {
				continue
			}
			h = append(h, head(x))
		}
	}
	return h, r
}

type jfloat float64

func jfloats32(j []jfloat) []float32 {
	r := make([]float32, len(j))
	for i, f := range j {
		r[i] = float32(f)
	}
	return r
}
func jsemis(j []jfloat) []int32 {
	r := make([]int32, len(j))
	for i, f := range j {
		r[i] = f.semi()
	}
	return r
}
func (j jfloat) semi() int32 {
	if math.IsNaN(float64(j)) {
		return 0x7FFFFFFF
	}
	return int32(math.Pow(2, 31) * float64(j) / 180.0)
}

func (f *jfloat) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "-123456" {
		*f = jfloat(math.NaN())
		return nil
	}
	n, e := strconv.ParseFloat(s, 64)
	*f = jfloat(n)
	return e
}
