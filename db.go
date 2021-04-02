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
}

type DiskDB struct {
	dir   string
	index []Header
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
func (d DiskDB) indexpath() string      { return filepath.Join(d.dir, "index.txt") }
func (d DiskDB) filepath(f File) string { return filepath.Join(d.dir, strconv.FormatInt(f.Start, 10)) }
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
func Each(db DB, g func(i int, f File)) {
	for i := 0; i < db.Len(); i++ {
		f, e := db.File(i)
		if e == nil {
			g(i, f)
		} else {
			fmt.Fprintln(os.Stderr, e)
		}
	}
}
func EachH(db DB, g func(i int, h Header)) {
	for i := 0; i < db.Len(); i++ {
		g(i, db.Head(i))
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
	return d, nil
}

type SingleFile File

func (s SingleFile) Len() int                 { return 1 }
func (s SingleFile) Head(i int) Header        { return s.Header }
func (s SingleFile) File(i int) (File, error) { return File(s), nil }

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
