package justeyecenters

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"

	"github.com/bamiaux/rez"
	"github.com/disintegration/gift"
	"gonum.org/v1/gonum/mat"
)

var resizer rez.Converter
var gausser = gift.New(gift.GaussianBlur(3.5))

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
	}, float64(255)*0.9)
	sobelY := convolve(resizedMat, [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}, float64(255)*0.9)

	gaussed := image.NewGray(resized.Bounds())
	gausser.Draw(gaussed, resized)
	gaussedMat := imageGray2Mat(gaussed, sizeX, sizeY)

	results := objective(gaussedMat, sobelX, sobelY, sizeX, sizeY)

	resultImg := mat2Image(results)
	var outBuf bytes.Buffer
	err = jpeg.Encode(&outBuf, resultImg, nil)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile("heatmap.jpg", outBuf.Bytes(), 0644)
	if err != nil {
		return nil, err
	}

	finalX, finalY := argmax2D(results)

	return &image.Point{
		X: finalX * (maxBounds.X / sizeX),
		Y: finalY * (maxBounds.Y / sizeY),
	}, nil
}

func mat2Image(m mat.Matrix) image.Image {
	sizeY, sizeX := m.Dims()
	img := image.NewGray(image.Rect(0, 0, sizeX, sizeY))

	maxX, maxY := argmax2D(m)
	maxValue := m.At(maxX, maxY)

	for y := 0; y < sizeY; y++ {
		for x := 0; x < sizeX; x++ {
			nextValue := m.At(x, y)
			nextValue *= 255 / maxValue
			img.Set(x, y, color.Gray{Y: uint8(nextValue)})
		}
	}

	return img
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
	for cY := 0; cY < sizeY; cY++ {
		for cX := 0; cX < sizeX; cX++ {
			dX, dY := makeUnitDisplacementMats(cX, cY, sizeX, sizeY)
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
