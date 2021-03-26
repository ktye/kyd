package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

// kyd -add -fit pqwerl.fit # returns timestamp
// kyd -table

func main() {
	var add, table bool
	var fit string
	var span string
	var dir string
	flag.BoolVar(&add, "add", false, "add/import")
	flag.BoolVar(&table, "table", false, "print file as table")
	flag.StringVar(&span, "date", "", "time span 2020.09.12-2020.08.17 or year")
	flag.StringVar(&dir, "dir", "./db/", "db directory")
	flag.StringVar(&fit, "fit", "", "fit file")
	flag.Parse()

	var db DB
	if fit != "" {
		f, e := ReadFit(fit)
		fatal(e)
		db = SingleFile(f)
	} else {
		var e error
		db, e = OpenDB(dir)
		fatal(e)
	}

	if span != "" {
		start, end := parseSpan(span)
		db = Range(db, start, end)
	}

	if add {
		_, o := db.(SingleFile)
		if o == false {
			panic("add: no file")
		}
		fmt.Println("todo")
	} else if table {
		for i := 0; i < db.Len(); i++ {
			f, e := db.File(i)
			fatal(e)
			f.Table(os.Stdout)
		}
	}
	fmt.Println("no command")
}
func parseSpan(s string) (int64, int64) {
	if s == "" {
		return 0, math.MaxInt64
	}
	if len(s) == 4 { // 2021
		y, e := strconv.Atoi(s)
		if e != nil {
			fatal(fmt.Errorf("cannot parse %s", s))
		}
		start := time.Date(y, 1, 1, 0, 0, 0, 0, nil)
		end := start.AddDate(1, 0, 0)
		return start.Unix(), end.Unix()
	} else if len(s) == 7 { // 2021.01
		start, e := time.Parse("2006.01", s)
		fatal(e)
		return start.Unix(), start.AddDate(0, 1, 0).Unix()
	} else if len(s) == 10 { // 2021.02.29
		start, e := time.Parse("2006.01.02", s)
		fatal(e)
		return start.Unix(), start.AddDate(0, 0, 1).Unix()
	} else if len(s) == 21 { // 2021.02.28-2021.03.30
		start, e := time.Parse("2006.01.02", s[:10])
		fatal(e)
		end, e := time.Parse("2006.01.02", s[11:])
		fatal(e)
		return start.Unix(), end.Unix()
	}
	fatal(fmt.Errorf("cannot parse range: %s", s))
	return 0, 0
}
func fatal(e error) {
	if e != nil {
		panic(e)
	}
}
