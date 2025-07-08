/*
By Lazar Savkovic 300246649
*/

package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
)

type Point3D struct {
	X float64
	Y float64
	Z float64
}
type Plane3D struct {
	A float64
	B float64
	C float64
	D float64
}
type Plane3DwSupport struct {
	Plane3D
	SupportSize int
}

/*
Reads a file and returns a slice of points or the original point cloud
*/
func ReadXYZ(filename string) []Point3D {
	var points []Point3D
	content, err := os.Open(filename)
	if err != nil {
		return points
	}
	defer content.Close()
	Scanner := bufio.NewScanner(content)
	Scanner.Split(bufio.ScanWords)
	Scanner.Text() // the first three Scanner.Text() is to skip the XYZ text which cannot be converted to float
	Scanner.Text() // this is in hopes
	Scanner.Text()
	index := 0
	xyz := [3]float64{}
	for Scanner.Scan() {
		xyz[index%3], _ = strconv.ParseFloat(Scanner.Text(), 64)
		if index%3 == 2 {
			point := Point3D{xyz[0], xyz[1], xyz[2]}
			points = append(points, point)
		}

		//points = append(points, Point3D{X, Y, Z})
		index++
	}
	return points
}

/*
saves a slice of points to a file
*/
func SaveXYZ(filename string, points []Point3D) {
	newFile, err := os.Create(filename)
	if err != nil {
	}
	defer newFile.Close()
	for a := 0; a < len(points); a++ {
		values := fmt.Sprintf("%f", points[a].X)
		values += fmt.Sprintf("    %f", points[a].Y)
		values += fmt.Sprintf("    %f", points[a].Z)
		values += "\n"
		_, err = newFile.WriteString(values)
		if err != nil {
			newFile.Close()
		}
	}
}

func (p1 Plane3D) GetDistance(p2 Point3D) float64 {
	return math.Abs((p1.A*p2.X)+p1.B*p2.Y+p1.C*p2.Z+p1.D) / (math.Sqrt(math.Pow(p1.A, 2)) + math.Sqrt(math.Pow(p1.B, 2)) + math.Sqrt(math.Pow(p1.C, 2)))
}

func GetNumberOfIterations(confidence float64, percentageOfPointsOnPlane float64) int {
	return (int)(math.Log(1-confidence) / (math.Log(1 - math.Pow(percentageOfPointsOnPlane, 3))))
}

/*
gets the supporting points and helps to identify the dominant plane
*/
func GetSupport(plane Plane3D, points []Point3D, eps float64, c chan Plane3DwSupport) Plane3DwSupport {
	support := Plane3DwSupport{}
	support.Plane3D = plane
	for a := 0; a < len(points); a++ { // runs through all points
		if plane.GetDistance(points[a]) < eps { // checks to see if the distance is smaller than eps
			support.SupportSize++ // if smaller then eps add this to the supported points
		}
	}
	c <- support
	return support
}

/*
gets the supporting points and is executed after the most dominant plane is found
*/
func GetSupportingPoints(plane Plane3D, points []Point3D, eps float64) []Point3D {
	var arrayPoints []Point3D          // starts off empty as there are no points
	for a := 0; a < len(points); a++ { // runs through all points
		if plane.GetDistance(points[a]) < eps { // checks to see if the distance is smaller than eps
			arrayPoints = append(arrayPoints, points[a]) // if smaller then append to the slice of points
		}
	}
	return arrayPoints
}

/*
does the opposit of Get supporting points and gets all points that are not supported
*/
func RemovePlane(plane Plane3D, points []Point3D, eps float64) []Point3D {
	var arrayPoints []Point3D
	for a := 0; a < len(points); a++ { //goes through all points
		if plane.GetDistance(points[a]) > eps { // if greater than eps
			arrayPoints = append(arrayPoints, points[a])
		}
	}
	return arrayPoints
}

/*
returns a plane using three points to create and return a plane
*/
func GetPlane(points []Point3D) Plane3D {
	a := (points[1].Y-points[0].Y)*(points[2].Z-points[0].Z) - (points[2].Y-points[0].Y)*(points[1].Z-points[0].Z) //value of A
	b := (points[1].Z-points[0].Z)*(points[2].X-points[0].X) - (points[2].Z-points[0].Z)*(points[1].X-points[0].X) //value of B
	c := (points[1].X-points[0].X)*(points[2].Y-points[0].Y) - (points[2].Y-points[0].Y)*(points[1].Y-points[0].Y) //value of C
	d := -(a*points[0].X + b*points[0].Y + c*points[0].Z)
	plane := Plane3D{a, b, c, d}
	return plane
}

func main() {
	//fmt.Printf("%d", GetNumberOfIterations(0.99, 0.50))
	var points []Point3D
	points = ReadXYZ("PointCloud3_p2_p0.xyz")
	//SaveXYZ("d.txt", points)
	iterations := GetNumberOfIterations(0.99, 0.50)
	var Support Plane3DwSupport
	Support.SupportSize = 0
	supportPlaneReceiver := make(chan Plane3DwSupport)
	for i := 0; i < iterations; i++ {
		var threePoints []Point3D
		for generateThree := 0; generateThree < 3; generateThree++ {
			threePoints = append(threePoints, points[rand.Intn(len(points))])
		}
		plane := GetPlane(threePoints)
		go GetSupport(plane, points, 1.0, supportPlaneReceiver)
	}
	temp := Plane3DwSupport{}
	temp.SupportSize = 0
	for a := 0; a < iterations; a++ {
		temp := <-supportPlaneReceiver
		if Support.SupportSize < temp.SupportSize {
			Support = temp
		}
	}
	fmt.Printf("%v", Support.SupportSize)                                            // tells you the support size
	close(supportPlaneReceiver)                                                      // close all channels as we no longer need them
	SaveXYZ("PointCloud3_p3.xyz", GetSupportingPoints(Support.Plane3D, points, 1.0)) // change file names if you want to go all the files

	SaveXYZ("PointCloud3_p3_p0.xyz", RemovePlane(Support.Plane3D, points, 1.0)) //uncoment at the last point cloud or p3
}
