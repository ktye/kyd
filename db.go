package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type DB interface {
	Len() int
	Head(i int) Header
	File(i int) (File, error)
	Races() []Race
}

type DiskDB struct {
	dir   string
	index []Header
	races []Race
}

func (d DiskDB) Len() int          { return len(d.index) }
func (d DiskDB) Head(i int) Header { return d.index[i] }
func (d DiskDB) File(i int) (File, error) {
	b, e := ioutil.ReadFile(filepath.Join(d.dir, strconv.FormatInt(d.index[i].Start, 10)))
	if e != nil {
		return File{}, e
	}
	return Decode(b)
}
func (d DiskDB) Races() []Race          { return d.races }
func (d DiskDB) indexpath() string      { return filepath.Join(d.dir, "index.txt") }
func (d DiskDB) filepath(f File) string { return filepath.Join(d.dir, strconv.FormatInt(f.Start, 10)) }
func (d DiskDB) racepath() string       { return filepath.Join(d.dir, "race.txt") }
func (d DiskDB) Add(f File) error {
	for i := 0; i < d.Len(); i++ {
		if d.Head(i).Start == f.Start {
			return fmt.Errorf("%d: file already exists in index", f.Start)
		}
	}
	fp, e := os.OpenFile(d.indexpath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	defer fp.Close()
	_, e = fmt.Fprintln(fp, f.Indexline())
	if e != nil {
		return e
	}
	if f.Samples == 0 {
		return nil
	}
	w, e := os.Create(d.filepath(f))
	if e != nil {
		return e
	}
	defer w.Close()
	return f.Encode(w)
}

func FindH(db DB, id int64) (h Header, e error) {
	for i := 0; i < db.Len(); i++ {
		h := db.Head(i)
		if h.Start == id {
			return h, nil
		}
	}
	return Header{}, fmt.Errorf("id not found: %d", id)
}
func Find(db DB, id int64) (f File, e error) {
	for i := 0; i < db.Len(); i++ {
		h := db.Head(i)
		if h.Start == id {
			return db.File(i)
		}
	}
	return f, fmt.Errorf("id not found: %d", id)
}
func NextId(db DB, id int64, prev bool) int64 {
	n := db.Len()
	for i := 0; i < n; i++ {
		h := db.Head(i)
		if h.Start == id {
			if prev {
				if i == 0 {
					return id
				} else {
					return db.Head(i - 1).Start
				}
			} else {
				if i == n-1 {
					return id
				} else {
					return db.Head(i + 1).Start
				}
			}
		}
	}
	return 0 // not reached
}
func Each(db DB, g func(i int, f File)) {
	for i := 0; i < db.Len(); i++ {
		f, e := db.File(i)
		if e == nil {
			g(i, f)
		} else if os.IsNotExist(e) == false {
			fmt.Fprintf(os.Stderr, "%d: %s\n", f.Start, e)
		}
	}
}
func EachH(db DB, g func(i int, h Header)) {
	for i := 0; i < db.Len(); i++ {
		g(i, db.Head(i))
	}
}
func EachR(db DB, g func(i int, r Race)) {
	rc := db.Races()
	for i, r := range rc {
		g(i, r)
	}
}
func Totals(db DB) (n int, t time.Duration, km float64, samples uint64) {
	n = db.Len()
	for i := 0; i < db.Len(); i++ {
		h := db.Head(i)
		t += time.Duration(int64(h.Seconds)) * time.Second
		km += float64(h.Meters) / 1000
		samples += h.Samples
	}
	return
}
func Years(db DB) {
	R := make(map[int]float64)
	B := make(map[int]float64)
	H := make(map[int]time.Duration)
	y0, y1 := 3000, 0
	for i := 0; i < db.Len(); i++ {
		h := db.Head(i)
		y := unix(h.Start).Year()
		if y < y0 {
			y0 = y
		}
		if y > y1 {
			y1 = y
		}
		H[y] += time.Duration(int64(h.Seconds)) * time.Second
		if h.Type == 1 {
			R[y] += float64(h.Meters) / 1000
		} else if h.Type == 2 {
			B[y] += float64(h.Meters) / 1000
		}
	}
	fmt.Printf("year R/km B/km H\n")
	for y := y0; y <= y1; y++ {
		fmt.Printf("%d %4.0f %4.0f %3d\n", y, R[y], B[y], H[y]/time.Hour)
	}
}

func OpenDB(dir string) (DiskDB, error) {
	d := DiskDB{dir: dir}
	b, e := ioutil.ReadFile(d.indexpath())
	if e != nil {
		return DiskDB{}, e
	}
	s := bufio.NewScanner(bytes.NewReader(b))
	line := 0
	for s.Scan() {
		line++
		t := s.Text()
		if t == "" {
			continue
		}
		h, e := ParseHeader(t)
		if e != nil {
			return DiskDB{}, fmt.Errorf("%s:%d: %s", d.indexpath(), line, e)
		}
		d.index = append(d.index, h)
	}
	b, e = ioutil.ReadFile(d.racepath())
	fatal(e)
	r, e := ReadRaces(bytes.NewReader(b))
	fatal(e)
	d.races = r
	return d, nil
}

type SingleFile File

func (s SingleFile) Len() int                 { return 1 }
func (s SingleFile) Head(i int) Header        { return s.Header }
func (s SingleFile) File(i int) (File, error) { return File(s), nil }
func (s SingleFile) Races() []Race            { return nil }

func Filter(d DB, g func(f File) bool) SubDB {
	s := SubDB{d: d, m: make(map[int]int)}
	k := 0
	for i := 0; i < d.Len(); i++ {
		f, e := d.File(i)
		if e == nil && g(f) {
			s.m[k] = i
			k++
		}
	}
	return s
}
func FilterH(d DB, g func(h Header) bool) SubDB {
	s := SubDB{d: d, m: make(map[int]int)}
	k := 0
	for i := 0; i < d.Len(); i++ {
		if g(d.Head(i)) {
			s.m[k] = i
			k++
		}
	}
	return s
}
func DateFilter(start, end int64) func(h Header) bool {
	return func(h Header) bool {
		return h.Start >= start && h.Start <= end
	}
}

type SubDB struct {
	d DB
	m map[int]int
}

func (d SubDB) Len() int                 { return len(d.m) }
func (d SubDB) Head(i int) Header        { return d.d.Head(d.m[i]) }
func (d SubDB) File(i int) (File, error) { return d.d.File(d.m[i]) }
func (d SubDB) Races() []Race            { return nil }
