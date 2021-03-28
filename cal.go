package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"strings"
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

func (c Cal) Write(w io.Writer, html bool, hi int) {
	var o *bufio.Writer
	var b bytes.Buffer
	if html {
		o = bufio.NewWriter(w)
		w = &b
		if hi >= 0 {
			hi = len(c) - 1 - hi
		}
	}
	bar := func(x float64, c byte) string { return strings.Repeat(string(c), int(math.Round(x))) }
	var ids []int64
	var tip []string
	tw := tabwriter.NewWriter(w, 2, 0, 3, ' ', 0)
	th, tkm, trkm, tbkm := 0.0, 0.0, 0.0, 0.0
	for _, wk := range c {
		fmt.Fprintf(tw, "%04d/%02d\t", wk.Year, wk.Week)
		for i := 0; i < 7; i++ {
			s, id := links(wk.Day[i])
			ids = append(ids, id...)
			for _, h := range wk.Day[i] {
				date := time.Unix(h.Start, 0).Format("2006.01.02")
				tip = append(tip, fmt.Sprintf("%s %.0fkm %v", date, h.Meters/1000, time.Duration(h.Seconds)*time.Second))
			}
			fmt.Fprintf(tw, "%s\t", s)
		}
		h, km, rkm, bkm, hs, hr, hb := weekly(wk.Day[:])
		hist := bar(hs, 'X') + bar(hr, 'Y') + bar(hb, 'Z')
		th, tkm, trkm, tbkm = th+h, tkm+km, trkm+rkm, tbkm+bkm
		fmt.Fprintf(tw, "%.1f\t%.0f\t%.0f\t%.0f\t%s\n", h, km, rkm, bkm, hist)
	}
	fmt.Fprintf(tw, "%dwk\t\t\t\t\t\t\t\t", len(c))
	fmt.Fprintf(tw, "%.0fh\t%.0fkm\t%.0fr\t%.0fb\n", th, tkm, trkm, tbkm)
	tw.Flush()

	if html {
		o.Write([]byte(calHead))
		s := bufio.NewScanner(&b)
		k := 0
		f := func(s string, id int64, t string) {
			fmt.Fprintf(o, `<a class="%s" id="%d" title="%s">%s</a>`, s, id, t, s)
		}
		i := 0
		for s.Scan() {
			t := strings.Replace(s.Text(), " ", "&nbsp;", -1)
			if i == hi {
				t = `<a id="hi" class="hi">` + t[:7] + `</a>` + t[7:]
			}
			for _, c := range []byte(t) {
				switch c {
				case 'S':
					f("S", ids[k], tip[k])
					k++
				case 'R':
					f("R", ids[k], tip[k])
					k++
				case 'B':
					f("B", ids[k], tip[k])
					k++
				case 'X':
					o.WriteString(`<a class="S">█</a>`)
				case 'Y':
					o.WriteString(`<a class="B">█</a>`)
				case 'Z':
					o.WriteString(`<a class="R">█</a>`)
				default:
					o.WriteByte(c)
				}
			}
			o.WriteByte(10)
			i++
		}
		o.Write([]byte(calTail))
		o.Flush()
	}
}
func (c Cal) WriteStrip(w io.Writer) error { // png image 50xWeeks
	if len(c) == 0 {
		return fmt.Errorf("empty calendar")
	}
	m := image.NewRGBA(image.Rect(0, 0, 50, len(c)))
	y := 0
	for i := len(c) - 1; i >= 0; i-- {
		wk := c[i]
		var s, b, r float64
		for _, d := range wk.Day {
			for _, h := range d {
				switch h.Type {
				case 1:
					r += float64(h.Seconds)
				case 2:
					b += float64(h.Seconds)
				case 5:
					s += float64(h.Seconds)
				}
			}
		}
		x := 0
		draw := func(h float64, c color.RGBA) {
			for k := 0; k < int(math.Round(h/3600)); k++ {
				m.SetRGBA(x, y, c)
				x++
			}
		}
		draw(s, blue)
		draw(b, green)
		draw(r, red)
		y++
	}
	return png.Encode(w, m)
}
func links(heads []Header) (s string, ids []int64) {
	for _, h := range heads {
		s += string(sport(h.Type))
		ids = append(ids, h.Start)
	}
	return s, ids
}
func weekly(a [][]Header) (hours, km, Rkm, Bkm, hs, hr, hb float64) {
	for _, heads := range a {
		for _, h := range heads {
			t, d := float64(h.Seconds)/3600, float64(h.Meters)/1000
			hours += t
			km += d
			if h.Type == 1 {
				Rkm += d
				hr += t
			} else if h.Type == 2 {
				Bkm += d
				hb += t
			} else if h.Type == 5 {
				hs += t
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

const calHead = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>kyd</title>
<link rel=icon href='favicon.png' />
<style>
 html{font-family:monospace}
 .S{color:blue}
 .R{color:red}
 .B{color:green}
 .hi{background:purple;color:white}
</style>
</head><body>
<pre>`

const calTail = `</pre>

<script>
function gu(x) { return (new URL(document.location)).searchParams.get(x) } // or null
function pa(x) { var r=gu(x);return r?("&"+x+"="+r):"" }

var p=pa("tile")
var all = document.querySelectorAll('a')
for (var i=0; i<all.length; i++) {
 var id=all[i].id; var a=all[i]
 a.href = (id.length)?"map.html?id="+id+p:""
 if(id=="hi")a.href=""
}
</script>

</body></html>
`
