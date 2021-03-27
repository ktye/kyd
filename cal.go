package main

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"
)

type Cal []Week // calendar

type Week struct {
	YearWeek
	Day [7][]Header
}

type YearWeek struct{ Year, Week int }

func Calendar(db DB) (cal Cal) {
	m := make(map[YearWeek]int)
	n := db.Len()
	if n == 0 {
		return nil
	}
	t := time.Unix(db.Head(0).Start, 0)
	last := time.Unix(db.Head(n-1).Start, 0)
	for i := 0; ; i++ {
		y, w := t.ISOWeek()
		cal = append(cal, Week{YearWeek: YearWeek{y, w}})
		m[YearWeek{y, w}] = i
		if t.After(last) {
			break
		}
		t = t.AddDate(0, 0, 7)
	}
	for i := 0; i < n; i++ {
		h := db.Head(i)
		k := m[h.yearweek()]
		d := h.day()
		cal[k].Day[d] = append(cal[k].Day[d], h)
	}
	return cal
}

func (c Cal) Write(w io.Writer) {
	tw := tabwriter.NewWriter(w, 2, 0, 3, ' ', 0)
	th, tkm, trkm, tbkm := 0.0, 0.0, 0.0, 0.0
	for _, wk := range c {
		fmt.Fprintf(tw, "%04d/%02d\t", wk.Year, wk.Week)
		for i := 0; i < 7; i++ {
			fmt.Fprintf(tw, "%s\t", links(wk.Day[i], true))
		}
		h, km, rkm, bkm := weekly(wk.Day[:])
		th, tkm, trkm, tbkm = th+h, tkm+km, trkm+rkm, tbkm+bkm
		fmt.Fprintf(tw, "%.1f\t%.0f\t%.0f\t%.0f\n", h, km, rkm, bkm)
	}
	fmt.Fprintf(tw, "%dwk\t\t\t\t\t\t\t\t", len(c))
	fmt.Fprintf(tw, "%.0fh\t%.0fkm\t%.0f(R)\t%.0f(B)\n", th, tkm, trkm, tbkm)
	tw.Flush()
}

func links(heads []Header, text bool) (s string) {
	for _, h := range heads {
		s += string(sport(h.Type))
	}
	return s
}
func weekly(a [][]Header) (hours, km, Rkm, Bkm float64) {
	for _, heads := range a {
		for _, h := range heads {
			hours += float64(h.Seconds / 3600)
			d := float64(h.Meters / 1000)
			km += d
			if h.Type == 1 {
				Rkm += d
			} else if h.Type == 2 {
				Bkm += d
			}
		}
	}
	return
}

func (h Header) yearweek() YearWeek {
	y, w := time.Unix(h.Start, 0).ISOWeek()
	return YearWeek{y, w}
}
func (h Header) day() int { // 0..6 mon..sun
	n := int(time.Unix(h.Start, 0).Weekday() - 1)
	if n < 0 {
		n += 7
	}
	return n
}
