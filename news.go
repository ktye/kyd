package main

var News map[uint64]int64

func makenews(db DB) {
	News = make(map[uint64]int64) // level-13 dot cache
	Each(db, func(i int, f File) {
		u := f.WebMercator()
		w := make([]uint64, len(u)/2)
		for j := 0; j < len(u); j += 2 {
			w[j/2] = uint64(u[j]>>11)<<32 | uint64(u[1+j]>>11)
		}
		for _, h := range w {
			if News[h] == 0 {
				News[h] = f.Start // mark file start as first step on that point
			}
		}
	})
}

func getnews(f File) []int8 {
	u := f.WebMercator()
	w := make([]uint64, len(u)/2)
	for j := 0; j < len(u); j += 2 {
		w[j/2] = uint64(u[j]>>11)<<32 | uint64(u[1+j]>>11)
	}
	news := make([]int8, len(w))
	for i, x := range w {
		if News[x] == f.Start {
			news[i] = 1
		}
	}
	for i := range news { // remove single points (todo short streaks?)
		if news[i] == 1 && i > 7 && news[i-1] == 0 && i < len(news)-1 && news[1+i] == 0 {
			news[i] = 0
		}
	}
	return news
}
