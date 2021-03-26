package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
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
	b, e := ioutil.ReadFile(filepath.Join(d.dir, strconv.FormatInt(d.index[i].Start, 64)))
	if e != nil {
		return File{}, e
	}
	return Decode(b)
}

func OpenDB(dir string) (DB, error) {
	d := DiskDB{dir: dir}
	index := filepath.Join(d.dir, "index.txt")
	b, e := ioutil.ReadFile(index)
	if e != nil {
		return nil, e
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
			return nil, fmt.Errorf("%s:%d: %s", index, line, e)
		}
		d.index = append(d.index, h)
	}
	return d, nil
}

type SingleFile File

func (s SingleFile) Len() int                 { return 1 }
func (s SingleFile) Head(i int) Header        { return s.Header }
func (s SingleFile) File(i int) (File, error) { return File(s), nil }

func Range(d DB, start, end int64) DB {
	r := RangeDB{d: d, start: start, end: end, m: make(map[int]int)}
	k := 0
	for i := 0; i < d.Len(); i++ {
		h := d.Head(i)
		if h.Start >= start && h.Start <= end {
			r.m[k] = i
			k++
		}
	}
	return r
}

type RangeDB struct {
	d          DB
	start, end int64
	m          map[int]int
}

func (d RangeDB) Len() int                 { return len(d.m) }
func (d RangeDB) Head(i int) Header        { return d.d.Head(d.m[i]) }
func (d RangeDB) File(i int) (File, error) { return d.d.File(d.m[i]) }
