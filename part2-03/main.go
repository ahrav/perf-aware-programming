package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/goccy/go-json"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage requires at least 2 arguments: <main> <json-file> <optional: profile>")
	}

	filename := os.Args[1]
	var shouldProfile bool
	if len(os.Args) > 2 {
		shouldProfile = os.Args[2] == "profile"
	}
	start := time.Now()
	pairs, err := readGeoPairsFromFile(filename)
	if shouldProfile {

		fmt.Printf("readGeoPairsFromFile took %s\n", time.Since(start))
	}
	if err != nil {
		log.Fatal(err)
	}

	startCalc := time.Now()
	fmt.Printf("Average distance: %f\n", calcHaversineDistanceAvg(pairs))
	if shouldProfile {
		fmt.Printf("calcHaversineDistanceAvg took %s\n", time.Since(startCalc))
		fmt.Printf("Total time: %s\n", time.Since(start))
	}
}

// GeoPair represents a pair of points in 2D space.
type GeoPair struct {
	X1 float64 `json:"X1"`
	Y1 float64 `json:"Y1"`
	X2 float64 `json:"X2"`
	Y2 float64 `json:"Y2"`
}

// GeoPairsContainer represents a container for a slice of GeoPairs.
type GeoPairsContainer struct {
	Pairs []GeoPair `json:"pairs"`
}

func readGeoPairsFromFile(filename string) ([]GeoPair, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var container GeoPairsContainer
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&container); err != nil {
		return nil, err
	}

	return container.Pairs, nil
}

func calcHaversineDistanceAvg(pairs []GeoPair) float64 {
	var sum float64
	for _, pair := range pairs {
		sum += Haversine(pair.Y1, pair.X1, pair.Y2, pair.X2)
	}
	return sum / float64(len(pairs))
}

// Square returns the square of the input.
func Square(x float64) float64 {
	return math.Pow(x, 2)
}

// Radians converts degrees to radians.
func Radians(d float64) float64 {
	return d * math.Pi / 180
}

// EarthRadius is the radius of the earth in kilometers.
const EarthRadius = 6372.8 // km

// Haversine computes the great circle distance between two points on the Earth.
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := Radians(lat2 - lat1)
	dLon := Radians(lon2 - lon1)
	lat1 = Radians(lat1)
	lat2 = Radians(lat2)

	a := Square(math.Sin(dLat/2)) + math.Cos(lat1)*math.Cos(lat2)*Square(math.Sin(dLon/2))
	c := 2 * math.Asin(math.Sqrt(a))
	return EarthRadius * c
}
