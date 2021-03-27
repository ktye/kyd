package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
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
	EachHead(db, func(i int, h Header) { fmt.Fprintln(w, h.String()) })
}
