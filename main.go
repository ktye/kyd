package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

func main() {
	var add, list, race, cal, table, totals, serve, unics bool
	var id int64
	var date, dir, here, addr, fit, imprt, diff string
	flag.BoolVar(&add, "add", false, "add/import")
	flag.BoolVar(&list, "list", false, "print header")
	flag.BoolVar(&race, "race", false, "print races")
	flag.BoolVar(&cal, "cal", false, "print calendar")
	flag.BoolVar(&table, "table", false, "print file as table")
	flag.BoolVar(&totals, "totals", false, "print db totals")
	flag.StringVar(&here, "here", "", "lat,lon (have i been here before?)")
	flag.BoolVar(&serve, "serve", false, "run as http server")
	flag.BoolVar(&unics, "unix", false, "print id as date")
	flag.Int64Var(&id, "id", 0, "use single file id")
	flag.StringVar(&date, "date", "", "time span 2020.09.12-2020.08.17 or year or year.month")
	flag.StringVar(&dir, "dir", "./db/", "db directory")
	flag.StringVar(&addr, "http", "127.0.0.1:2021", "serve on this address")
	flag.StringVar(&fit, "fit", "", "fit file")
	flag.StringVar(&imprt, "import", "", "import old db")
	flag.StringVar(&diff, "diff", "", "compare fit dir against the db")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "github.com/ktye/kyd")
		flag.PrintDefaults()
	}
	flag.Parse()

	if imprt != "" {
		importDB(imprt, dir)
		return
	}
	if unics != false {
		fmt.Println(unix(id).Format("20060102T150405"))
		return
	}

	var db DB
	if fit != "" {
		f, e := ReadFit(fit)
		fatal(e)
		db = SingleFile(f)
	} else {
		var e error
		db, e = OpenDB(dir)
		fatal(e)
		if id != 0 {
			f, e := Find(db, id)
			fatal(e)
			db = SingleFile(f)
		}
	}

	if date != "" {
		start, end := parseSpan(date)
		db = FilterH(db, DateFilter(start, end))
	}
	if here != "" {
		db = Here(db, here)
	}
	if diff != "" {
		fitDiff(db, diff)
		return
	}

	if add {
		f, o := db.(SingleFile)
		if o == false {
			panic("add: no single file")
		}
		db, e := OpenDB(dir)
		fatal(e)
		fatal(db.Add(File(f)))
		fmt.Println("a", f.Start)
	} else if list {
		EachH(db, func(i int, h Header) { fmt.Println(h.String()) })
	} else if race {
		EachR(db, func(i int, r Race) { fmt.Println(r.String()) })
	} else if cal {
		Calendar(db).Write(os.Stdout, false, -1)
	} else if table {
		Each(db, func(i int, f File) { f.Table(os.Stdout) })
	} else if totals {
		n, t, km, samples := Totals(db)
		fmt.Printf("#%d %v %.0fkm %dsamples\n", n, t, km, samples)
	} else if serve {
		server(addr, db)

	} else {
		fmt.Println("no command")
	}
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
		start := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
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
