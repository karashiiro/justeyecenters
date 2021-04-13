package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"os"

	"github.com/karashiiro/justeyecenters"
)

func main() {
	files, err := os.ReadDir("example/eyes/")
	if err != nil {
		log.Fatal(err)
	}

	for _, de := range files {
		f, err := os.Open("example/eyes/" + de.Name())
		if err != nil {
			log.Fatal(err)
		}

		img, err := jpeg.Decode(f)
		if err != nil {
			log.Fatal(err)
		}

		center, err := justeyecenters.GetEyeCenter(img)
		if err != nil {
			log.Fatal(err)
		}

		bounds := img.Bounds()
		imgCopy := image.NewRGBA(img.Bounds())
		for y := 0; y < bounds.Max.Y; y++ {
			for x := 0; x < bounds.Max.X; x++ {
				if math.Abs(float64(center.X-x)) <= 2 && math.Abs(float64(center.Y-y)) <= 2 {
					imgCopy.Set(x, y, color.RGBA{R: 255})
					continue
				}

				imgCopy.Set(x, y, img.At(x, y))
			}
		}

		var out bytes.Buffer
		err = jpeg.Encode(&out, imgCopy, nil)
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile("example/results/"+de.Name(), out.Bytes(), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
