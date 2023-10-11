package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
)

const mb = 1024 * 1024

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage requires at least 2 arguments: <main> <json-file> <optional: profile>")
	}

	var (
		shouldProfile bool
		shouldTest    bool
	)
	if len(os.Args) > 2 {
		shouldProfile = os.Args[2] == "profile"
		shouldTest = os.Args[2] == "test"
	}

	if shouldTest {
		repetitionTester(os.Args[1])
		return
	}

	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Get the file's size.
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	fileSize := info.Size()

	start := time.Now()
	pairs, err := readGeoPairsFromFile(file)
	if shouldProfile {
		dur := time.Since(start)
		throughput := float64(fileSize) / dur.Seconds() / mb
		fmt.Printf("readGeoPairsFromFile took %s throughput %f MB/s\n", dur, throughput)
	}
	if err != nil {
		log.Fatal(err)
	}

	startCalc := time.Now()
	fmt.Printf("Average distance: %f\n", calcHaversineDistanceAvg(pairs))
	if shouldProfile {
		dur := time.Since(startCalc)
		throughput := float64(fileSize) / dur.Seconds() / mb
		fmt.Printf("calcHaversineDistanceAvg took %s throughput %f MB/s\n", dur, throughput)
		fmt.Printf("Total time: %s\n", time.Since(start))
	}
}

const testDuration = 10 * time.Second

func repetitionTester(filename string) {

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return
	}
	defer file.Close()

	var (
		minTime, maxTime time.Duration
		iterationCount   int
	)
	resetTimer := true
	overallStart := time.Now()
	deadline := overallStart.Add(testDuration)

	for resetTimer || time.Now().Before(deadline) {
		resetTimer = false
		iterationCount++

		// Reset file position.
		file.Seek(0, 0)

		// Measure the time taken to execute readGeoPairsFromFile.
		start := time.Now()
		_, err := readGeoPairsFromFile(file)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("Error in iteration %d: %s\n", iterationCount, err)
			return
		}

		// Update min and max.
		if iterationCount == 1 || duration < minTime {
			minTime = duration
			deadline = time.Now().Add(testDuration) // Reset timer if new minimum found
			resetTimer = true
		}
		if iterationCount == 1 || duration > maxTime {
			maxTime = duration
		}
	}

	// Get the file's size.
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	fileSize := info.Size()

	fmt.Printf("Total iterations: %d\n", iterationCount)
	fmt.Printf("Minimum execution time: %v throughputs %f MB/s\n", minTime, float64(fileSize)/minTime.Seconds()/mb)
	fmt.Printf("Maximum execution time: %v throughputs %f MB/s\n", maxTime, float64(fileSize)/maxTime.Seconds()/mb)
}

func getPageFaults(pid int) (int, error) {
	// Run 'ps' command to get page faults (majflt/minflt).
	cmd := exec.Command("ps", "-o", "majflt,minflt", "-p", strconv.Itoa(pid))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	// Parse the output.
	lines := strings.Split(out.String(), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected output from ps command")
	}

	fields := strings.Fields(lines[1])
	if len(fields) != 2 {
		return 0, fmt.Errorf("unexpected field count from ps command")
	}

	// Assuming you want minor page faults, but you can return major (fields[0]) if needed.
	pageFaults, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, err
	}

	return pageFaults, nil
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

const expectedGeoPairs = 10_000_000 // Expected number of GeoPairs

func readGeoPairsFromFile(file *os.File) ([]GeoPair, error) {
	var container GeoPairsContainer
	container.Pairs = make([]GeoPair, 0, expectedGeoPairs)

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&container); err != nil {
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

// Square returns the square of the input.
func Square(x float64) float64 {
	return x * x
}

const reciprocal180 = 1.0 / 180

// Radians converts degrees to radians.
func Radians(d float64) float64 {
	return d * math.Pi * reciprocal180
}
