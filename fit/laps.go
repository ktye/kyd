//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/ascii85"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/tormoder/fit"
)

func main() {
	a := os.Args[1:]
	if len(a) == 0 {
		fatal(fmt.Errorf("laps file.fit..\nktye/kyd/fit/laps.go"))
	}
	for _, file := range a {
		F(file)
	}
}
func fatal(e error) {
	if e != nil {
		panic(e)
	}
}

var itoa = strconv.Itoa

func F(file string) {
	b, e := ioutil.ReadFile(file)
	fatal(e)

	var t *fit.File
	t, e = fit.Decode(bytes.NewReader(b))
	if e != nil && t == nil {
		return
	}

	var a *fit.ActivityFile
	a, e = t.Activity()
	if e != nil {
		return
	}

	if len(a.Sessions) == 0 {
		return
	}
	var time, dist uint32
	for _, s := range a.Sessions {
		time += s.TotalTimerTime // ms
		dist += s.TotalDistance  // 0.1m
	}

	if len(a.Laps) < 4 {
		return
	}
	t0 := a.Sessions[0].Timestamp

	var o []byte
	o = append(o, byte(len(a.Laps)))
	for _, l := range a.Laps {
		s := uint16(math.Round(0.001 * float64(l.TotalTimerTime)))
		d := uint16(math.Round(0.01 * float64(l.TotalDistance)))
		o = append(o, byte(d>>8), byte(d&0xff), byte(s>>8), byte(s&0xff))
	}

	rec := a.Records
	if len(rec) < 2 {
		return
	}

	b = nil
	for _, re := range rec {
		b = append(b, byte(re.HeartRate))
	}

	b = unfill(b)
	o = append(o, b[0])

	if len(b)&1 == 0 {
		b = append(b, b[len(b)-1])
	}

	b = deltas(b)

	o = append(o, b[0])
	b = b[1:]
	for i := 0; i < len(b); i += 2 {
		o = append(o, (b[i]<<8)|b[1+i])
	}

	fmt.Println(t0.Format("20060102") + ":" + b85(o) + ",")

	//fmt.Println(b)
	//let un85=s=>{let o=[],i=0,c,b,j,k,L,p,v;while(i<s.length){c=s[i++];if(c=="z"){o.push(0,0,0,0);continue};b=c;for(j=0;j<4&&i<s.length;j++)b+=s[i++];L=b.length;if(L<5)b+="u".repeat(5-L);v=0;for(k=0;k<5;k++){v=v*85+(b.charCodeAt(k)-33)};p=[(v>>>24)&255,(v>>>16)&255,(v>>>8)&255,v&255];for(let m=0;m<(L-1);m++)o.push(p[m])};return new Uint8Array(o)}
}

func unfill(b []byte) (r []byte) {
	l := byte(100)
	for _, c := range b {
		if c == 255 {
			c = l
		}
		r = append(r, c)
		l = c
	}
	return r
}
func deltas(b []byte) (r []byte) {
	r = append(r, b[0])
	x := int(b[0])
	for i := 1; i < len(b); i++ {
		d := int(b[i]) - x
		if d < -8 {
			d = -8
		} else if d > 7 {
			d = 7
		}
		x += d
		r = append(r, byte(d+8))
	}
	return r
}
func b85(b []byte) string {
	r := make([]byte, ascii85.MaxEncodedLen(len(b)))
	n := ascii85.Encode(r, b)
	r = r[:n]
	for i := range r {
		r[i] += 2
	}
	return "\"" + strings.ReplaceAll(string(r), "\\", "\\\\") + "\""
}
