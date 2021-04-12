package justeyecenters

import (
	"image"
	"math"

	"github.com/bamiaux/rez"
	"gonum.org/v1/gonum/mat"
)

var resizer rez.Converter

func GetEyeCenter(img image.Image) (*image.Point, error) {
	maxBounds := img.Bounds().Max
	gray := image.NewGray(img.Bounds())
	for x := 0; x < maxBounds.X; x++ {
		for y := 0; y < maxBounds.Y; y++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	resized := image.NewGray(image.Rect(0, 0, 64, 64))
	if resizer == nil {
		err := initResizer(resized, gray)
		if err != nil {
			return nil, err
		}
	}

	resizedBounds := resized.Bounds().Max
	sizeX := resizedBounds.X
	sizeY := resizedBounds.Y

	err := resizer.Convert(resized, gray)
	if err != nil {
		return nil, err
	}

	resizedMat := imageGray2Mat(resized, sizeX, sizeY)

	sobelX := convolve(resizedMat, [][]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	})
	sobelY := convolve(resizedMat, [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	})

	results := objective(resizedMat, sobelX, sobelY, sizeX, sizeY)

	finalX, finalY := argmax2D(results)

	return &image.Point{
		X: finalX * (maxBounds.X / sizeX),
		Y: finalY * (maxBounds.Y / sizeY),
	}, nil
}

func argmax2D(m mat.Matrix) (int, int) {
	sizeY, sizeX := m.Dims()
	lastMax := float64(0)
	lastMaxX := 0
	lastMaxY := 0

	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			nextValue := m.At(x, y)
			if nextValue >= lastMax {
				lastMax = nextValue
				lastMaxX = x
				lastMaxY = y
			}
		}
	}

	return lastMaxX, lastMaxY
}

func objective(gray, gradX, gradY mat.Matrix, sizeX, sizeY int) mat.Matrix {
	results := mat.NewDense(sizeY, sizeX, nil)
	totalElements := float64(sizeX * sizeY)
	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			dX, dY := makeUnitDisplacementMats(x, y, sizeX, sizeY)
			nextValue := float64(0)
			weight := 255 - gray.At(x, y)
			for cY := 0; cY < sizeY; cY++ {
				for cX := 0; cX < sizeX; cX++ {
					nextGradX := gradX.At(cX, cY)
					nextGradY := gradY.At(cX, cY)
					if nextGradX == 0 && nextGradY == 0 {
						continue
					}

					mag := math.Sqrt(nextGradX*nextGradX + nextGradY*nextGradY)

					prod := dX.At(cX, cY)*(nextGradX/mag) + dY.At(cX, cY)*(nextGradY/mag)
					nextValue += prod * prod * weight
				}
			}
			results.Set(x, y, nextValue/totalElements)
		}
	}
	return results
}

func convolve(inMat mat.Matrix, kernel [][]float64) mat.Matrix {
	rows, cols := inMat.Dims()
	kernelRows := len(kernel)
	kernelCols := len(kernel[0])
	output := mat.NewDense(rows, cols, nil)
	for y := 1; y < rows-1; y++ {
		for x := 1; x < rows-1; x++ {
			convResult := float64(0)
			for kY := -1; kY < kernelRows-1; kY++ {
				for kX := -1; kX < kernelCols-1; kX++ {
					convResult += inMat.At(x+kX, y+kY) * kernel[kX+1][kY+1]
				}
			}
			output.Set(x, y, convResult)
		}
	}
	return output
}

func imageGray2Mat(img image.Image, sizeX, sizeY int) mat.Matrix {
	output := mat.NewDense(sizeY, sizeX, nil)
	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			val, _, _, _ := img.At(x, y).RGBA()
			output.Set(x, y, float64(val>>8))
		}
	}
	return output
}

func makeUnitDisplacementMats(fromX, fromY, sizeX, sizeY int) (mat.Matrix, mat.Matrix) {
	outputX := mat.NewDense(sizeY, sizeX, nil)
	outputY := mat.NewDense(sizeY, sizeX, nil)
	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			if x == fromX || y == fromY {
				continue
			}

			dX := float64(x - fromX)
			dY := float64(y - fromY)
			mag := math.Sqrt(dX*dX + dY*dY)
			outputX.Set(x, y, dX/mag)
			outputY.Set(x, y, dY/mag)
		}
	}
	return outputX, outputY
}

func initResizer(output, input image.Image) error {
	cfg, err := rez.PrepareConversion(output, input)
	if err != nil {
		return err
	}

	resizer, err = rez.NewConverter(cfg, rez.NewBilinearFilter())
	if err != nil {
		return err
	}

	return nil
}
