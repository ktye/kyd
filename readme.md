track and point

# create db
```sh
mkdir db
touch db/index.txt
kyd -add -fit file.fit
```

# commands
## list
`kyd -list -date 2019`

## calendar (one week per line)
`kyd -cal`

## dump file
```sh
kyd -table -id 1394964105
kyd -table -date 2021.03.27
kyd -table -date 2021
```

## have i been here before?
`kyd -here 60.422018,7.184887`

# server
`kyd -serve [-http=$ADDR]`

## http api
```
/cal?w=       calendar (highlight week)
/head?id=..   header(text)
/json?id=..   File as json
/ll?id=..     lat lon(json)
/list  ?n= &s= &w= &e=   (query rectangle north/south/west/east)
/map.html?id=..             interactive map track over opentopmap
/map.html?tile=..id=.. generate tiles from all points in db (tile=points|grey|inferno)
/strip.png    barplot weekly hours, one pixel row per week
/tile/$z/$x/$y.png    tile server
```

## map multi-stage race/tour
```
/map.html?id=1459095710,1529758181,1536037659
```

## rectangular selection
shift+mouse drag in `map.html` draws a rectangle and creates a link (rect).
clicking on the like shows a list of all id's that contain points within the rectangle.

# database
the db is stored in a directory (default -db="./db/").
- `db/index.txt` text file, one entry per line (type Header)
- `db/race.txt` text file, one entry per line (type Race)
- `db/1394964105` binary file (name/id is unix seconds) (type File)

```go
type Header struct {
	Start   int64   // unix time (seconds)
	Type    uint32  // type 1(run) 2(cycle)
	Seconds float32 // total duration
	Meters  float32 // total distance
	Samples uint64  // number of samples
}
type File struct {
	Header
	Time []float32 // seconds
	Dist []float32 // meters
	Alt  []float32 // altitude (m)
	Lat  []int32   // semicircles (invalid: 0x7FFFFFFF) (180 / math.Pow(2, 31))
	Lon  []int32   // semicircles
}
type Race struct {
	Start  int64         // unix time (seconds)
	Type   string        // "800m"
	Time   time.Duration //
	Result string        // "101/2048"
	Name   string
}
```
