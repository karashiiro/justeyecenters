package justeyecenters_test

import (
	"bytes"
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

	newImg, err := justeyecenters.GetEyeCenter(img)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, newImg, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("testeye_new.jpg", buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}
}
