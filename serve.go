package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

//go:embed www
var www embed.FS

var db hdb
var root fs.FS
var tile Tile

type hdb struct {
	sync.Mutex
	DB
	cal Cal
}

func server(addr string, a DB) {
	db = hdb{DB: a, cal: Calendar(a)}
	tile = NewTile(db)
	fmt.Println(addr+"/index.html", len(tile.run)+len(tile.bike))

	var e error
	root, e = fs.Sub(www, "www")
	fatal(e)
	http.Handle("/", http.FileServer(http.FS(root)))
	template.ParseFS(www, "*.tmpl")
	http.HandleFunc("/index.html", serveIndex)
	http.HandleFunc("/strip.png", serveStrip)
	http.HandleFunc("/cal", serveCal)
	http.HandleFunc("/list", serveList)
	http.HandleFunc("/head", serveHead)
	http.HandleFunc("/json", serveJson)
	http.HandleFunc("/alt", serveAlt)
	http.HandleFunc("/ll", serveLatLon)
	http.HandleFunc("/tile/", serveTile)

	fatal(http.ListenAndServe(addr, nil))
}
func templ(w io.Writer, file string, data interface{}) {
	template.Must(template.ParseFS(root, file)).Execute(w, data)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	n, t, km, samples := Totals(db)
	totals := fmt.Sprintf("#%d %v %.0fkm %dsamples\n", n, t, km, samples)
	templ(w, "index.tmpl", totals)
}
func serveCal(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	wk := -1
	if i, e := strconv.Atoi(r.URL.Query().Get("w")); e == nil {
		wk = i
	}
	db.cal.Write(w, true, wk)
}
func serveList(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	var d DB = db
	if g := getRect(r); g != nil {
		d = Filter(db, g)
	}
	tile := r.URL.Query().Get("tile")
	type t struct {
		Id      int64
		Tile, S string
	}
	var heads []t
	EachHead(d, func(i int, h Header) { heads = append(heads, t{h.Start, tile, h.String()[11:]}) })
	templ(w, "list.tmpl", heads)
}
func getRect(r *http.Request) func(f File) bool {
	p := func(s string) float64 {
		n, e := strconv.ParseFloat(s, 64)
		if e != nil {
			return math.NaN()
		}
		return n
	}
	q := r.URL.Query()
	n, s, w, e := p(q.Get("n")), p(q.Get("s")), p(q.Get("w")), p(q.Get("e"))
	if math.IsNaN(n) || math.IsNaN(s) || math.IsNaN(w) || math.IsNaN(e) {
		return nil
	}
	return func(f File) bool {
		for i := uint64(0); i < f.Samples; i++ {
			lat, lon := Deg(f.Lat[i]), Deg(f.Lon[i])
			if lat <= n && lat >= s && lon >= w && lon <= e {
				return true
			}
		}
		return false
	}
}
func pa(r *http.Request, p string) string {
	v := r.URL.Query().Get(p)
	if v != "" {
		return "&" + p + "=" + v
	}
	return ""
}
func getId(r *http.Request) int64 {
	id, e := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if e != nil {
		log.Println(e)
		return 0
	}
	return id
}
func getHeader(r *http.Request) (Header, error) {
	db.Lock()
	defer db.Unlock()
	h, e := FindH(db, getId(r))
	if e != nil {
		fmt.Println(e)
	}
	return h, e
}
func getFile(r *http.Request) (File, error) {
	db.Lock()
	defer db.Unlock()
	f, e := Find(db, getId(r))
	if e != nil {
		fmt.Println(e)
	}
	return f, e
}
func serveHead(w http.ResponseWriter, r *http.Request) {
	h, e := getHeader(r)
	if e == nil {
		fmt.Fprintln(w, h.String())
	}
}
func serveJson(w http.ResponseWriter, r *http.Request) {
	f, e := getFile(r)
	if e == nil {
		json.NewEncoder(w).Encode(f)
	}
}
func serveAlt(w http.ResponseWriter, r *http.Request) {
	W, H := 600, 50
	f, e := getFile(r)
	if e == nil && f.Samples > 0 {
		w.Header().Set("Content-Type", "image/png")
		m := image.NewRGBA(image.Rect(0, 0, W, H))
		max32 := func(x []float32) float64 {
			m := 0.0
			for _, v := range x {
				if u := float64(v); math.IsNaN(u) == false && u > m {
					m = u
				}
			}
			return m
		}
		xs, ys := 0.001, 0.1
		dm, am := max32(f.Dist), max32(f.Alt)
		for xs*dm > float64(W) {
			xs /= 2
		}
		for ys*am > float64(H) {
			ys /= 2
		}
		for i := uint64(0); i < f.Samples; i++ {
			x, y := xs*float64(f.Dist[i]), float64(H)-ys*float64(f.Alt[i])
			if math.IsNaN(x) == false && math.IsNaN(y) == false {
				m.Set(int(x), int(y), color.Black)
			}
		}
		png.Encode(w, m)
	}
}
func serveLatLon(w http.ResponseWriter, r *http.Request) {
	f, e := getFile(r)
	if e == nil {
		p := make([][2]float64, 0, f.Samples)
		for i := uint64(0); i < f.Samples; i++ {
			la, lo := Deg(f.Lat[i]), Deg(f.Lon[i])
			if math.IsNaN(la) == false && math.IsNaN(lo) == false {
				p = append(p, [2]float64{la, lo})
			}
		}
		if e := json.NewEncoder(w).Encode(p); e != nil {
			fmt.Println("ll", e)
		}
	} else {
		fmt.Println("ll", e)
	}
}
func serveTile(w http.ResponseWriter, r *http.Request) {
	v := strings.Split(r.URL.Path, "/") // /tile/11/1023/234.png
	if len(v) != 5 {
		fmt.Println("tile: wrong path:", r.URL.Path)
	}
	p := func(s string) uint32 {
		u, e := strconv.ParseUint(s, 10, 32)
		if e != nil {
			fmt.Println("tile:", e)
		}
		return uint32(u)
	}
	v[4] = strings.TrimSuffix(v[4], ".png")

	w.Header().Set("Content-Type", "image/png")
	tile.Png(w, p(v[2]), p(v[3]), p(v[4]))
}
func serveStrip(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	w.Header().Set("Content-Type", "image/png")
	if e := db.cal.WriteStrip(w); e != nil {
		fmt.Println(e)
	}
}
