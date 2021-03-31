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

func (t Tile) Png(w io.Writer, z, x, y uint32, tileType string) {
	z = 24 - z
	d := uint32(256 << z)
	x0, y0 := x*d, y*d

	var newImage func() image.Image
	var drwImage func(image.Image) image.Image
	if tileType == "grey" {
		newImage = func() image.Image { return image.NewGray(image.Rect(0, 0, 256, 256)) }
		drwImage = func(m image.Image) image.Image {
			im := m.(*image.Gray)
			draw := func(u []uint32) {
				for i := 0; i < len(u); i += 2 {
					if x, y := (u[i]-x0)>>z, (u[i+1]-y0)>>z; x < 256 && y < 256 {
						g := im.GrayAt(int(x), int(y))
						if c := g.Y; c < 20 {
							g.Y = 20
						} else if c < 255 {
							g.Y++
						}
						im.SetGray(int(x), int(y), g)
					}
				}
			}
			draw(t.bike)
			draw(t.run)
			return im
		}
	} else if tileType == "inferno" {
		var p [256][256]byte
		newImage = func() image.Image { return image.NewRGBA(image.Rect(0, 0, 256, 256)) }
		drwImage = func(m image.Image) image.Image {
			im := m.(*image.RGBA)
			draw := func(u []uint32) {
				for i := 0; i < len(u); i += 2 {
					if x, y := (u[i]-x0)>>z, (u[i+1]-y0)>>z; x < 256 && y < 256 {
						if c := p[int(x)][int(y)]; c < 255 {
							p[int(x)][int(y)] = 1 + c
						}
					}
				}
				for i := 0; i < 256; i++ {
					for k := 0; k < 256; k++ {
						im.SetRGBA(i, k, inferno[p[i][k]])
					}
				}
			}
			draw(t.bike)
			draw(t.run)
			return im
		}
	} else {
		newImage = func() image.Image { return image.NewRGBA(image.Rect(0, 0, 256, 256)) }
		drwImage = func(m image.Image) image.Image {
			im := m.(*image.RGBA)
			draw := func(u []uint32, c color.RGBA) {
				for i := 0; i < len(u); i += 2 {
					if x, y := (u[i]-x0)>>z, (u[i+1]-y0)>>z; x < 256 && y < 256 {
						im.SetRGBA(int(x), int(y), c)
					}
				}
			}
			draw(t.bike, green)
			draw(t.run, red)
			return im
		}
	}
	png.Encode(w, drwImage(newImage()))
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

var (
	red   = color.RGBA{230, 57, 0, 255}
	green = color.RGBA{0, 230, 115, 255}
	blue  = color.RGBA{0, 71, 171, 255}
)
