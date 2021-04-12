package justeyecenters

import (
	"image"

	"github.com/bamiaux/rez"
)

var resizer *rez.Converter

func GetEyeCenter(img image.Image) (*image.Point, error) {
	resized := image.NewGray(image.Rect(0, 0, 64, 64))
	if resizer == nil {
		err := initResizer(resized, img)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func initResizer(output, input image.Image) error {
	cfg, err := rez.PrepareConversion(output, input)
	if err != nil {
		return err
	}

	cvt, err := rez.NewConverter(cfg, rez.NewBilinearFilter())
	if err != nil {
		return err
	}

	resizer = &cvt

	return nil
}
