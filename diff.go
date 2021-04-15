package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func fitDiff(db DB, dir string) {
	files, e := os.ReadDir(dir)
	fatal(e)
	dh := 0
	for _, fi := range files {
		if fi.IsDir() == false {
			name := filepath.Join(dir, fi.Name())
			f, e := ReadFit(name)
			if e != nil {
				fmt.Println(name, e)
			}
			fatal(e)
			//fmt.Println(f.Start, f.Samples, name)
			g, e := Find(db, f.Start)
			if e != nil {
				g, e = Find(db, f.Start+3600)
				dh++
			}
			if e != nil {
				fmt.Println(f.Start, f.Samples, name, e)
			} else {
				e = diffFile(name, f, g)
				if e != nil {
					fmt.Println(f.Start, f.Samples, name, e)
				}
			}
		}
	}
	fmt.Println("dh:", dh)
}

func diffFile(name string, f, g File) error {
	if f.Start != g.Start && f.Start+3600 != g.Start {
		return fmt.Errorf("%s: Start %v %v", name, f.Start, g.Start)
	}
	if f.Type != g.Type {
		return fmt.Errorf("%s: Type %v %v", name, f.Type, g.Type)
	}
	if math.Abs(float64(round(f.Seconds)-g.Seconds)) > 1 {
		return fmt.Errorf("%s: Seconds %v %v", name, f.Seconds, g.Seconds)
	}
	if f.Meters != g.Meters {
		return fmt.Errorf("%s: Meters %v %v", name, f.Meters, g.Meters)
	}
	if f.Samples != g.Samples {
		return fmt.Errorf("%s: Samples %v %v", name, f.Samples, g.Samples)
	}
	for i := uint64(0); i < f.Samples; i++ {
		if diffFloats(f.Time[i], g.Time[i]) {
			return fmt.Errorf("%s: Time[%d] %v %v", name, i, f.Time[i], g.Time[i])
		}
		if diffFloats(f.Dist[i], g.Dist[i]) {
			return fmt.Errorf("%s: Dist[%d] %v %v", name, i, f.Dist[i], g.Dist[i])
		}
		if diffFloats(round(f.Alt[i]), g.Alt[i]) {
			return fmt.Errorf("%s: Alt[%d] %v %v", name, i, f.Alt[i], g.Alt[i])
		}
		if diffLL(f.Lat[i], g.Lat[i]) {
			return fmt.Errorf("%s: Lat[%d] %v %v", name, i, f.Lat[i], g.Lat[i])
		}
		if diffLL(f.Lon[i], g.Lon[i]) {
			return fmt.Errorf("%s: Lon[%d] %v %v", name, i, f.Lon[i], g.Lon[i])
		}
	}
	return nil
}
func round(a float32) float32 { return float32(math.Round(float64(a))) }
func diffFloats(a, b float32) bool {
	if math.IsNaN(float64(a)) && math.IsNaN(float64(b)) {
		return false
	}
	return a != b
}
func diffLL(a, b int32) bool {
	if a-b < -2 || a-b > 2 {
		return true
	}
	return false
}
