package main

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
)

type Tile struct {
	run, bike []uint32 // lat,lon full-range web mercator projection
}

func NewTile(db DB) (t Tile) {
	Each(db, func(i int, f File) {
		if f.Type == 1 {
			t.run = append(t.run, f.WebMercator()...)
		} else if f.Type == 2 {
			t.bike = append(t.bike, f.WebMercator()...)
		}
	})
	return t
}

func (t Tile) Png(w io.Writer, z, x, y uint32) {

	/*
		s := fmt.Sprintf("tile %d/%d/%d.png points:", z, x, y)
		var points int
		defer func() {
			fmt.Println(s, points)
		}()
	*/

	z = 24 - z
	d := uint32(256 << z)
	x0, y0 := x*d, y*d

	m := image.NewRGBA(image.Rect(0, 0, 256, 256))
	u := t.bike
	c := color.RGBA{0, 230, 115, 255} // bike(green)

	draw := func() {
		for i := 0; i < len(u); i += 2 {
			if x, y := (u[i]-x0)>>z, (u[i+1]-y0)>>z; x < 256 && y < 256 {
				m.Set(int(x), int(y), c)
				//points++
			}
		}
	}
	draw()
	u = t.run
	c = color.RGBA{230, 57, 0, 255} // run(red)
	draw()
	png.Encode(w, m)
}

func (f File) WebMercator() []uint32 {
	p := make([]uint32, 0, 2*f.Samples)
	for i := uint64(0); i < f.Samples; i++ {
		if a, b, o := mercator(f.Lat[i], f.Lon[i]); o {
			p = append(p, a, b)
		}
	}
	return p
}

// semicirles to web-mercator (full range)
func mercator(lat, lon int32) (x uint32, y uint32, ok bool) {
	ok = lat != invalidSemis && lon != invalidSemis
	if !ok {
		return
	}
	const s = float64(math.MaxUint32 / 2)
	la := rad(Deg(lat))
	x = uint32(lon + math.MinInt32)
	y = uint32(s * (1 - math.Log(math.Tan(la)+1/math.Cos(la))/math.Pi))
	//fmt.Println(lat, lon, Deg(lat), Deg(lon), x, y)
	return x, y, la < 1.4844 && la > -1.4844
}
