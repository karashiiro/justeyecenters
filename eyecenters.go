package justeyecenters

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/bamiaux/rez"
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

	err := resizer.Convert(resized, gray)
	if err != nil {
		return nil, err
	}

	_ = effect.Sobel(resized)

	return &image.Point{
		X: maxBounds.X / 2,
		Y: maxBounds.Y / 2,
	}, nil
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
