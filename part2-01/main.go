package main

import (
	"encoding/binary"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
)

type SampleType string

const (
	Uniform   SampleType = "uniform"
	Clustered SampleType = "clustered"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("usage requires 4 arguments: <main> <uniform/clustered> <random-seed> <num-points>")
	}

	spread := os.Args[1]
	if spread != string(Uniform) && spread != string(Clustered) {
		log.Fatal("usage requires 4 arguments: <main> <uniform/clustered> <random-seed> <num-points>")
	}

	seed, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	numPoints, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}

	// Generate the data.
	if spread == string(Uniform) {
		uniform(seed, numPoints)
	} else {
		clustered(seed, numPoints)
	}
}

// Point represents a point in 2D space.
type Point struct {
	X, Y float64
}

// uniform generates a uniform distribution of haversine distances.
func uniform(seed, numPoints int) {
	// Set the random seed.
	r := rand.New(rand.NewSource(int64(seed)))

	binFile, err := os.Create("data.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer binFile.Close()

	outputFile, err := os.Create("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()

	// Output file json format:
	// { "pairs": [
	// {"X1": <float>, "Y1": <float>, "X2": <float>, "Y2": <float>, "distance": <float>},
	// ...
	// ]}
	outputFile.WriteString("{\"pairs\": [\n")

	sum := 0.0
	// Generate a uniform distribution of haversine distances.
	for i := 0; i < numPoints; i++ {
		p1 := Point{r.Float64()*360 - 180, r.Float64()*180 - 90}
		p2 := Point{r.Float64()*360 - 180, r.Float64()*180 - 90}
		dist := Haversine(p1.Y, p1.X, p2.Y, p2.X)
		if err := binary.Write(binFile, binary.LittleEndian, dist); err != nil {
			log.Fatal(err)
		}
		sum += dist

		// For the iterations after the first, append a comma before the data.
		if i > 0 {
			outputFile.WriteString(",\n")
		}

		// Write the data (without a comma at the end)
		dataString := "{\"X1\": " + strconv.FormatFloat(p1.X, 'f', -1, 64) +
			", \"Y1\": " + strconv.FormatFloat(p1.Y, 'f', -1, 64) +
			", \"X2\": " + strconv.FormatFloat(p2.X, 'f', -1, 64) +
			", \"Y2\": " + strconv.FormatFloat(p2.Y, 'f', -1, 64) +
			", \"distance\": " + strconv.FormatFloat(dist, 'f', -1, 64) + "}"
		if _, err := outputFile.WriteString(dataString); err != nil {
			log.Fatal(err)
		}
	}
	outputFile.WriteString("\n]}\n")

	// Print the average distance.
	log.Printf("Average distance: %f", sum/float64(numPoints))
	log.Println("Number of points:", numPoints)
	log.Println("Random seed:", seed)
}

// NumClusters is the number of clusters to generate.
const (
	NumClusters int     = 32
	ClusterSize float64 = 32
)

type Cluster struct {
	Xmin, Xmax float64
	Ymin, Ymax float64
}

// clustered generates a clustered distribution of haversine distances.
func clustered(seed, numPoints int) {
	r := rand.New(rand.NewSource(int64(seed)))

	// Generate a clustered distribution of points.
	ptsPerCluster := numPoints / NumClusters
	clusters := make([]Cluster, NumClusters)
	for i := 0; i < NumClusters; i++ {
		// Random center for each cluster
		centerX := r.Float64()*360 - 180 // Longitude between -180 and 180
		centerY := r.Float64()*180 - 90  // Latitude between -90 and 90

		// Create a bounding box around the center.
		clusters[i] = Cluster{
			Xmin: math.Max(centerX-ClusterSize, -180),
			Xmax: math.Min(centerX+ClusterSize, 180),
			Ymin: math.Max(centerY-ClusterSize, -90),
			Ymax: math.Min(centerY+ClusterSize, 90),
		}
	}

	binFile, err := os.Create("data.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer binFile.Close()

	outputFile, err := os.Create("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	// Output file json format:
	// { "pairs": [
	// {"X1": <float>, "Y1": <float>, "X2": <float>, "Y2": <float>, "distance": <float>},
	// ...
	// ]}
	outputFile.WriteString("{\"pairs\": [\n")

	type result struct {
		sum  float64
		data string
	}

	resultsChan := make(chan result, len(clusters))
	wg := sync.WaitGroup{}

	for idx, c := range clusters {
		wg.Add(1)
		go func(idx int, c Cluster) {
			defer wg.Done()
			var localSum float64
			var dataBuilder strings.Builder

			// Create a local unique random generator for this goroutine since rand is not thread-safe.
			src := rand.NewSource(int64(seed) + int64(idx))
			lr := rand.New(src)

			for j := 0; j < ptsPerCluster; j++ {
				p1 := Point{X: lr.Float64()*(c.Xmax-c.Xmin) + c.Xmin, Y: lr.Float64()*(c.Ymax-c.Ymin) + c.Ymin}
				p2 := Point{X: lr.Float64()*(c.Xmax-c.Xmin) + c.Xmin, Y: lr.Float64()*(c.Ymax-c.Ymin) + c.Ymin}
				dist := Haversine(p1.Y, p1.X, p2.Y, p2.X)
				localSum += dist

				dataString := "{\"X1\": " + strconv.FormatFloat(p1.X, 'f', -1, 64) +
					", \"Y1\": " + strconv.FormatFloat(p1.Y, 'f', -1, 64) +
					", \"X2\": " + strconv.FormatFloat(p2.X, 'f', -1, 64) +
					", \"Y2\": " + strconv.FormatFloat(p2.Y, 'f', -1, 64) + "}"

				dataBuilder.WriteString(dataString)
				// Add comma if not last cluster or not last point within cluster.
				// Only add a comma if it's not the last point within the cluster
				if j != ptsPerCluster-1 {
					dataBuilder.WriteString(",\n")
				}
			}

			resultsChan <- result{sum: localSum, data: dataBuilder.String()}
		}(idx, c)
	}

	// Close resultsChan once all goroutines are done.
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Read from resultsChan.
	var globalSum float64
	firstResult := true
	for res := range resultsChan {
		if err := binary.Write(binFile, binary.LittleEndian, res.sum); err != nil {
			log.Fatal(err)
		}
		globalSum += res.sum

		if firstResult {
			firstResult = false
		} else {
			outputFile.WriteString(",\n") // Prefix with comma if not the first result.
		}

		if _, err := outputFile.WriteString(res.data); err != nil {
			log.Fatal(err)
		}
	}
	outputFile.WriteString("\n]}\n")

	// Print the average distance.
	log.Printf("Average distance: %f", globalSum/float64(numPoints))
	log.Println("Number of points:", numPoints)
	log.Println("Random seed:", seed)

	return
}

func clusteredSerial(seed, numPoints int) {
	// Generate a clustered distribution of points.
	ptsPerCluster := numPoints / NumClusters
	var clusters []Cluster
	stepX := 360.0 / float64(NumClusters) // Longitude steps
	stepY := 180.0 / float64(NumClusters) // Latitude steps

	for i := 0; i < NumClusters; i++ {
		// Each cluster will only contain points within a certain range of latitudes and longitudes.
		// The range of the lattiudes and longitudes will be different for each cluster depending on the cluster number.
		clusters = append(clusters, Cluster{
			Xmin: -180 + stepX*float64(i),
			Xmax: -180 + stepX*float64(i+1),
			Ymin: -90 + stepY*float64(i),
			Ymax: -90 + stepY*float64(i+1),
		})
	}

	r := rand.New(rand.NewSource(int64(seed)))
	binFile, err := os.Create("data.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer binFile.Close()

	outputFile, err := os.Create("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	// Output file json format:
	// { "pairs": [
	// {"X1": <float>, "Y1": <float>, "X2": <float>, "Y2": <float>, "distance": <float>},
	// ...
	// ]}
	outputFile.WriteString("{\"pairs\": [\n")

	sum := 0.0
	totalClusters := len(clusters)
	for i, c := range clusters {
		// Generate a uniform distribution of haversine distances.
		for j := 0; j < ptsPerCluster; j++ {
			p1 := Point{X: r.Float64()*(c.Xmax-c.Xmin) + c.Xmin, Y: r.Float64()*(c.Ymax-c.Ymin) + c.Ymin}
			p2 := Point{X: r.Float64()*(c.Xmax-c.Xmin) + c.Xmin, Y: r.Float64()*(c.Ymax-c.Ymin) + c.Ymin}
			dist := Haversine(p1.Y, p1.X, p2.Y, p2.X)
			if err := binary.Write(binFile, binary.LittleEndian, dist); err != nil {
				log.Fatal(err)
			}
			sum += dist

			// Write the data (without a comma at the end)
			dataString := "{\"X1\": " + strconv.FormatFloat(p1.X, 'f', -1, 64) +
				", \"Y1\": " + strconv.FormatFloat(p1.Y, 'f', -1, 64) +
				", \"X2\": " + strconv.FormatFloat(p2.X, 'f', -1, 64) +
				", \"Y2\": " + strconv.FormatFloat(p2.Y, 'f', -1, 64) + "}"

			if _, err := outputFile.WriteString(dataString); err != nil {
				log.Fatal(err)
			}

			// If it's not the last iteration of both loops, append a comma.
			if i != totalClusters-1 || j != ptsPerCluster-1 {
				outputFile.WriteString(",\n")
			}
		}
	}
	outputFile.WriteString("\n]}\n")

	// Print the average distance.
	log.Printf("Average distance: %f", sum/float64(numPoints))
	log.Println("Number of points:", numPoints)
	log.Println("Random seed:", seed)

	return
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
