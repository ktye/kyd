package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func serveBitmap(addr string, db DB, a []string) {
	er := fmt.Errorf("usage: kyd bitmap lat lon zoom")
	if len(a) == 2 {
		a = append(a, "15")
	}
	if len(a) != 3 {
		fatal(er)
	}
	la, lo, ok := mercator(semis(parseFloat(a[0]), parseFloat(a[1])))
	if !ok {
		fatal(er)
	}
	zoom, e := strconv.Atoi(a[2])
	zoom = 24 - zoom
	fatal(e)
	la = la>>zoom - 512
	lo = lo>>zoom - 512

	m := make(map[uint32]bool)
	q := make([]int32, 0)
	push := func(lat, lon int32) {
		x, y, ok := mercator(lat, lon)
		if !ok {
			return
		}
		x = x>>zoom - la
		y = y>>zoom - lo
		if x < 1024 && y < 1024 {
			i := 1024*x + y
			if m[i] == false {
				m[i] = true
				q = append(q, int32(i))
			}
		}
	}

	n := db.Len()
	for i := 0; i < n; i++ {
		if h := db.Head(i); h.Samples == 0 {
			continue
		}
		f, e := db.File(i)
		if e != nil {
			panic(e)
		}
		for j := range f.Lat {
			push(f.Lat[j], f.Lon[j])
		}
	}
	fmt.Printf("points: %d (%.1f%%)\n", len(q), float64(100*len(q))/float64(1024*1024))

	b, e := json.Marshal(q)
	fatal(e)

	http.HandleFunc("/x", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(bitmapHtml))
	})
	fmt.Println(addr + "/")
	fatal(http.ListenAndServe(addr, nil))
}
func semis(lat, lon float64) (int32, int32) {
	z := float64((1 << 31) / 180)
	return int32(z * lat), int32(z * lon)
}

const bitmapHtml = `<!DOCTYPE html>
<head><meta charset="utf-8"><title></title>
<style>*{background:black}</style>
</head>
<body onload="init()">
<canvas id="cnv" width="1024" height="1024"></canvas>
<script>
function init(){
 fetch("x").then(r=>r.json()).then(r=>{
  let c=document.getElementById("cnv").getContext("2d")
  //c.fillStyle="black";c.fillRect(0,0,1024,1024);
  c.fillStyle="white"
  let i=0;
  let dot=()=>{
   c.fillRect(r[  i]>>10,r[  i]&1023,1,1)
   c.fillRect(r[1+i]>>10,r[1+i]&1023,1,1)
   c.fillRect(r[2+i]>>10,r[2+i]&1023,1,1)
   i+=3;if(i<r.length-2)window.requestAnimationFrame(dot)
  }
  window.requestAnimationFrame(dot)
 })
}
</script>
</body>
</html>
`
