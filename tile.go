package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
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

	fmt.Println("tile", z, x, y)

	d := uint32(256 << z)
	x0, y0 := x*d, y*d

	m := image.NewRGBA(image.Rect(0, 0, 128, 128))
	u := t.bike
	c := color.RGBA{0, 0, 255, 255}

	draw := func() {
		for i := 0; i < len(u); i += 2 {
			if x, y := u[i], u[i+1]; x >= x0 && x < x0+d && y >= y0 && y < y0+d {
				x, y = (x-x0)>>z, (y-y0)>>z
				m.Set(int(x), int(y), c)
			}
		}
	}
	draw()
	u = t.run
	c = color.RGBA{255, 0, 0, 255}
	draw()
	png.Encode(w, m)
}
