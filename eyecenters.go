package justeyecenters

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/bamiaux/rez"
)

var resizer rez.Converter

func GetEyeCenter(img image.Image) (image.Image, error) {
	resized := image.NewRGBA(image.Rect(0, 0, 64, 64))
	if resizer == nil {
		err := initResizer(resized, img)
		if err != nil {
			return nil, err
		}
	}

	err := resizer.Convert(resized, img)
	if err != nil {
		return nil, err
	}

	sobelized := effect.Sobel(resized)

	return sobelized, nil
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
