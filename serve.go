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
	"sync"
)

//go:embed www
var www embed.FS

var db hdb
var root fs.FS

type hdb struct {
	sync.Mutex
	DB
	cal Cal
}

func server(addr string, a DB) {
	fmt.Println(addr + "/index.html")

	db = hdb{DB: a, cal: Calendar(a)}

	var e error
	root, e = fs.Sub(www, "www")
	fatal(e)
	http.Handle("/", http.FileServer(http.FS(root)))
	template.ParseFS(www, "*.tmpl")
	http.HandleFunc("/index.html", serveIndex)
	http.HandleFunc("/cal", serveCal)
	http.HandleFunc("/list", serveList)
	http.HandleFunc("/head", serveHead)
	http.HandleFunc("/json", serveJson)
	http.HandleFunc("/alt", serveAlt)
	http.HandleFunc("/ll", serveLatLon)

	fatal(http.ListenAndServe(addr, nil))
}
func templ(w io.Writer, file string, data interface{}) {
	template.Must(template.ParseFS(root, file)).Execute(w, data)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	//q := r.URL.Query()
	//date := q.Get("date")

	n, t, km, samples := Totals(db)
	totals := fmt.Sprintf("#%d %v %.0fkm %dsamples\n", n, t, km, samples)
	templ(w, "index.tmpl", totals)
}
func serveCal(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	db.cal.Write(w)
}
func serveList(w http.ResponseWriter, r *http.Request) {
	db.Lock()
	defer db.Unlock()
	var d DB = db
	if g := getRect(r); g != nil {
		d = Filter(db, g)
	}
	EachHead(d, func(i int, h Header) { fmt.Fprintln(w, h.String()) })
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
	f, e := getFile(r)
	if e == nil && f.Samples > 0 {
		w.Header().Set("Content-Type", "image/png")
		m := image.NewRGBA(image.Rect(0, 0, 600, 100))
		for i := uint64(0); i < f.Samples; i++ {
			x, y := float64(f.Dist[i])/1000, float64(f.Alt[i])
			x, y = 2*x, (1000-y)/10
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
		p := make([][2]float64, f.Samples)
		for i := uint64(0); i < f.Samples; i++ {
			p[i][0], p[i][1] = Deg(f.Lat[i]), Deg(f.Lon[i])
		}
		json.NewEncoder(w).Encode(p)
	}
}
