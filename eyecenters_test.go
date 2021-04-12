package justeyecenters_test

import (
	"image"
	"image/jpeg"
	"os"
	"testing"

	"github.com/karashiiro/justeyecenters"
)

func TestGetEyeCenter(t *testing.T) {
	f, err := os.Open("testeye.jpg")
	if err != nil {
		t.Fatal(err)
	}

	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	prediction, err := justeyecenters.GetEyeCenter(img)
	if err != nil {
		t.Fatal(err)
	}

	expected := image.Rect(98, 54, 121, 77)
	if !prediction.In(expected) {
		t.Fatalf("predicted center not in expected bounds. prediction: %v; bounds: %v", prediction, expected)
	}
}
