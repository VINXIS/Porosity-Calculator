package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"regexp"
	"strconv"
)

// Create a struct with 4 fields: sample#, iteration#, direction, and porosity
// Direction is either H or V
type Image struct {
	sampleNum int
	iterNum   int
	porosity  float64
	direction string
	b         float64
}

type refImage struct {
	name string
	img  image.Image
}

// Take an image and calculate the percentage of the image that is black
func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Obtain the original folder directory
	// Iterate over the images in the original folder
	// For each image, change the images to black and white and save the image to the processed folder
	// For each image, calculate the percentage of the image that is black and save the percentage to a file

	// Obtain the original folder directory
	dir := "./original"
	originalFolder, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer originalFolder.Close()

	// Iterate over the images in the original folder
	names, err := originalFolder.Readdirnames(-1)
	if err != nil {
		log.Fatal(err)
	}

	// Create an array of refImages from names
	log.Println("Loading images")
	var refImgs []refImage
	for _, name := range names {
		// Get the path to the image
		imgPath := fmt.Sprintf(dir+"/%s", name)

		// Load the image
		img := loadImg(imgPath)

		// Create a refImage struct
		refImg := refImage{
			name: name,
			img:  img,
		}

		// Add the refImage to the array
		refImgs = append(refImgs, refImg)
	}

	log.Println("Finished loading all images. Starting to process images with b = 0 to 255...")

	imgs := make(chan Image, 256*len(refImgs))

	for i := 0.0; i <= 255; i++ {
		// Create the b value folder if it doesn't exist
		bFolder := fmt.Sprintf("./processed/%d", int(i))
		err := os.MkdirAll(bFolder, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		go calcPorosity(i, refImgs, dir, imgs)
	}

	// Put all the images from the channel into one array
	log.Println("Finished processing all images, writing to csv file")

	// Create a csv file for porosity
	f, err := os.Create("./processed/porosity.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Create a csv writer
	w := csv.NewWriter(f)

	// Write the headers to the csv file
	err = w.Write([]string{"Sample Number", "Iteration Number", "Porosity", "Direction", "B"})
	if err != nil {
		log.Fatal(err)
	}

	// Get images from channel and write them to the csv file
	for i := 0; i < 256*len(refImgs); i++ {
		img := <-imgs

		// Write the porosity data to the csv file
		err = w.Write([]string{strconv.Itoa(img.sampleNum), strconv.Itoa(img.iterNum), strconv.FormatFloat(img.porosity, 'f', 2, 64), img.direction, strconv.FormatFloat(img.b, 'f', 0, 64)})
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Finished writing", img.sampleNum, img.iterNum, img.direction, img.b, "to csv files")
	}

	// Flush the writers
	w.Flush()
}

// func calcPorosity takes in a b value, an array of refImages, and the directory of the original images containing the images to be processed and calculates the porosity of each image
func calcPorosity(b float64, refImgs []refImage, dir string, imgs chan Image) {
	for _, refImg := range refImgs {

		// Change the image to black and white
		newImg := blackwhite(refImg.img, b)

		// For each image, calculate the percentage of the image that is black and note the file name associate to the percentage in the porosity text file
		blackPercent := calcBlackPercent(newImg)

		// Save the image to the processed folder under the b value folder
		go func() {
			bFolder := fmt.Sprintf("./processed/%d", int(b))

			// Save the image to the b value folder
			imgPath := fmt.Sprintf(bFolder+"/%s", refImg.name)
			saveImg(newImg, imgPath)
		}()

		// The file name is saved in the form x-yD, where x is the sample number, y is the iteration number, and D is the direction (H or V)

		// Create regex to find the sample, iteration, and direction
		re := regexp.MustCompile(`(\d+)-(\d+)([HV])`)
		matches := re.FindStringSubmatch(refImg.name)

		// Get the sample number
		sampleNum, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatal(err)
		}

		// Get the iteration number
		iterNum, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Fatal(err)
		}

		// Get the direction
		direction := matches[3]

		// Create an Image struct
		img := Image{
			sampleNum: sampleNum,
			iterNum:   iterNum,
			porosity:  blackPercent,
			direction: direction,
			b:         b,
		}

		// Add the Image struct to the array
		imgs <- img

		log.Println("Finished processing", refImg.name, "with b =", b)
	}
}

// func loadImg takes the path to an image (jpg or png) and returns an image.Image
func loadImg(path string) image.Image {
	// Open the image
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()

	return img
}

// func saveImg takes an image.Image and saves it to the specified path
func saveImg(img image.Image, path string) {
	// Create the file
	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	// Encode the image
	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()
}

// func blackwhite takes a jpg/png image in the form of an image.Image and returns a black and white image in the form of an image.Image
func blackwhite(img image.Image, thresh float64) image.Image {
	// Create a new image with the same dimensions as the original image
	newImg := image.NewRGBA(img.Bounds())

	// Iterate over the pixels in the image
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			// Get the pixel at (x, y)
			pixel := img.At(x, y)

			// Convert the pixel to a color.RGBA
			_, _, b, a := pixel.RGBA()

			// Check if the pixel reaches the threshold of black
			if float64(b)/float64(a) <= thresh/255.0 {
				// Set the pixel to black
				newImg.Set(x, y, color.RGBA{0, 0, 0, 255})

			} else {
				// Set the pixel to white
				newImg.Set(x, y, color.RGBA{255, 255, 255, 255})

			}
		}
	}

	return newImg
}

// func calcBlackPercent takes an image.Image and returns a float64
func calcBlackPercent(img image.Image) float64 {
	// Create a variable to hold the number of black pixels
	var blackPixels int

	// Iterate over the pixels in the image
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			// Get the pixel at (x, y)
			pixel := img.At(x, y)

			// Convert the pixel to a color.RGBA
			r, g, b, _ := pixel.RGBA()

			// Check if the pixel is black
			if r == 0 && g == 0 && b == 0 {
				blackPixels++
			}
		}
	}

	// Calculate the percentage of the image that is black
	blackPercent := float64(blackPixels) / float64(img.Bounds().Dx()*img.Bounds().Dy()) * 100

	return blackPercent
}
