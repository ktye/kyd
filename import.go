package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func importDB(dir string) {
	h, r := importIndex(dir)
	fmt.Println(len(h), len(r))
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

	seconds := func(s string) float32 {
		if s == "DNF" || s == "" {
			return float32(math.NaN())
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
		return float32(d.Seconds())
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
			Start:   fileTime(x.File),
			Type:    x.Racetype,
			Seconds: seconds(x.Racetime),
			Result:  x.Result,
			Name:    x.Title,
		}
	}
	head := func(x hdr) Header {
		if x.Dist == "" {
			fmt.Println(x)
		}
		return Header{
			Start:   fileTime(x.File),
			Type:    parseType(x.Type),
			Seconds: seconds(x.Time),
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
