package justeyecenters

import (
	"image"
	"math"

	"github.com/bamiaux/rez"
	"github.com/disintegration/gift"
	"gonum.org/v1/gonum/mat"
)

var resizer rez.Converter
var gausser = gift.New(gift.GaussianBlur(3.5))

var sobelXKernel = [][]float64{
	{-1, 0, 1},
	{-2, 0, 2},
	{-1, 0, 1},
}

var sobelYKernel = [][]float64{
	{-1, -2, -1},
	{0, 0, 0},
	{1, 2, 1},
}

// GetEyeCenter predicts an eye center location based on a cropped input
// image. The input image should be cropped to just fit the eye; significant
// deviations from these bounds will reduce the accuracy of the predictor
// dramatically.
//
// Implemented based on Timm, F. and Barth, E. (2011). "Accurate eye centre
// localisation by means of gradients".
func GetEyeCenter(img image.Image) (*image.Point, error) {
	maxBounds := img.Bounds().Max

	// Convert to grayscale
	gray := image.NewGray(img.Bounds())
	for x := 0; x < maxBounds.X; x++ {
		for y := 0; y < maxBounds.Y; y++ {
			gray.Set(x, y, img.At(x, y))
		}
	}

	// Downscale the image to 32x32
	resized := image.NewGray(image.Rect(0, 0, 32, 32))
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

	// Blur the downscaled image
	gaussed := image.NewGray(resized.Bounds())
	gausser.Draw(gaussed, resized)
	gaussedMat := imageGray2Mat(gaussed, sizeX, sizeY)

	// Generate the image gradient
	resizedMat := imageGray2Mat(resized, sizeX, sizeY)
	sobelX := convolve(resizedMat, sobelXKernel, float64(255)*0.9)
	sobelY := convolve(resizedMat, sobelYKernel, float64(255)*0.9)

	// Run the objective function
	results := objective(gaussedMat, sobelX, sobelY, sizeX, sizeY)

	// Get final center estimate
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

	// Preallocate displacement matrices since their values aren't reused
	// between loop iterations
	dX := mat.NewDense(sizeY, sizeX, nil)
	dY := mat.NewDense(sizeY, sizeX, nil)

	for cY := 0; cY < sizeY; cY++ {
		for cX := 0; cX < sizeX; cX++ {
			makeUnitDisplacementMats(dX, dY, cX, cY, sizeX, sizeY)
			nextValue := float64(0)
			weight := 255 - gray.At(cX, cY)
			for y := 0; y < sizeY; y++ {
				for x := 0; x < sizeX; x++ {
					nextGradX := gradX.At(x, y)
					nextGradY := gradY.At(x, y)
					if nextGradX == 0 && nextGradY == 0 {
						continue
					}

					mag := math.Sqrt(nextGradX*nextGradX + nextGradY*nextGradY)

					prod := dX.At(x, y)*(nextGradY/mag) + dY.At(x, y)*(nextGradX/mag)
					nextValue += prod * prod * weight
				}
			}
			results.Set(cX, cY, nextValue/totalElements)
		}
	}

	return results
}

func convolve(inMat mat.Matrix, kernel [][]float64, threshold float64) mat.Matrix {
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

			if math.Abs(convResult) < threshold {
				continue
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

func makeUnitDisplacementMats(mX, mY *mat.Dense, fromX, fromY, sizeX, sizeY int) {
	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			if x == fromX || y == fromY {
				mX.Set(x, y, 0)
				mY.Set(x, y, 0)
				continue
			}

			dX := float64(x - fromX)
			dY := float64(y - fromY)
			mag := math.Sqrt(dX*dX + dY*dY)
			mX.Set(x, y, dX/mag)
			mY.Set(x, y, dY/mag)
		}
	}
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
