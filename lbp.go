package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"image"
	"image/color"
	"image/jpeg"
)

type Lbp struct {
	image   []uint8
	dim     int
	mapping []uint8
}

func NewLbp() *Lbp {
	lbp := new(Lbp)
	lbp.dim = 58
	lbp.mapping = make([]uint8, 256)

	for i := 0; i < 256; i++ {
		lbp.mapping[i] = 57
	}

	lbp.mapping[0x00] = 56
	lbp.mapping[0xFF] = 56

	var i, j uint8
	for i = 0; i < 8; i++ {
		for j = 0; j < 8; j++ {
			s := (1 << j) - 1
			s <<= i
			s = (s | (s >> 8)) & 0xFF

			lbp.mapping[s] = i*7 + (j - 1)
		}
	}

	return lbp
}

func (lbp *Lbp) preprocess(im image.Image, bounds image.Rectangle) {
	lbp.image = make([]uint8, bounds.Max.Y*bounds.Max.X)

	// Store the image as a Y * X uint8 array (grayscale)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(im.At(x, y)).(color.Gray)
			lbp.image[x+y*bounds.Max.X] = gray.Y
		}
	}
}

func (lbp *Lbp) at(x int, y int, w int) uint8 {
	return lbp.image[x+y*w]
}

func (lbp *Lbp) Process(im image.Image, regionSize int) []float64 {
	bounds := im.Bounds()
	lbp.preprocess(im, bounds)

	rwidth := bounds.Max.X / regionSize
	rheight := bounds.Max.Y / regionSize
	rstride := rwidth * rheight

	features := make([]float64, lbp.dim*rstride)

	var bits uint8
	var wy1, wy2, wx1, wx2 float64
	maxX := bounds.Max.X
	for y := 1; y < bounds.Max.Y-1; y++ {

		wy1 = (float64(y)+0.5)/float64(regionSize) - 0.5
		ry1 := int(math.Floor(wy1))
		ry2 := ry1 + 1
		wy2 = wy1 - float64(ry1)
		wy1 = 1.0 - wy2

		if ry1 >= rheight {
			continue
		}

		for x := 1; x < bounds.Max.X-1; x++ {
			wx1 = (float64(x)+0.5)/float64(regionSize) - 0.5
			rx1 := int(math.Floor(wx1))
			rx2 := rx1 + 1
			wx2 = wx1 - float64(rx1)
			wx1 = 1.0 - wx2

			if rx1 >= rwidth {
				continue
			}

			bits = 0
			center := lbp.at(x, y, maxX)
			if lbp.at(x+1, y, maxX) > center {
				bits |= 0x1 << 0 // E
			}
			if lbp.at(x+1, y+1, maxX) > center {
				bits |= 0x1 << 1 // SE
			}
			if lbp.at(x+0, y+1, maxX) > center {
				bits |= 0x1 << 2 // S
			}
			if lbp.at(x-1, y+1, maxX) > center {
				bits |= 0x1 << 3 // SW
			}
			if lbp.at(x-1, y+0, maxX) > center {
				bits |= 0x1 << 4 // W
			}
			if lbp.at(x-1, y-1, maxX) > center {
				bits |= 0x1 << 5 // NW
			}
			if lbp.at(x+0, y-1, maxX) > center {
				bits |= 0x1 << 6 // N
			}
			if lbp.at(x+1, y-1, maxX) > center {
				bits |= 0x1 << 7 // NW
			}

			bin := int(lbp.mapping[bits])

			if rx1 >= 0 && ry1 >= 0 {
				features[bin*rstride+ry1*rwidth+rx1] += wx1 * wy1
			}
			if rx2 < rwidth && ry1 >= 0 {
				features[bin*rstride+ry1*rwidth+rx2] += wx2 * wy1
			}
			if rx1 >= 0 && ry2 < rheight {
				features[bin*rstride+ry2*rwidth+rx1] += wx1 * wy2
			}
			if rx2 < rwidth && ry2 < rheight {
				features[bin*rstride+ry2*rwidth+rx2] += wx2 * wy2
			}

		}
	}

	// Normalize regions
	var offset int
	var norm float64
	for ry := 0; ry < rheight; ry++ {
		for rx := 0; rx < rwidth; rx++ {
			norm = 0
			for k := 0; k < lbp.dim; k++ {
				norm += features[k*rstride+offset]
			}
			norm = math.Sqrt(norm) + 1e-10
			for k := 0; k < lbp.dim; k++ {
				features[k*rstride+offset] = math.Sqrt(features[k*rstride+offset]) / norm
			}
		}
		offset++
	}

	return features
}

func main() {
	r, err := os.Open("Barack-Obama5.jpg")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer r.Close()

	im, err := jpeg.Decode(r)
	if err != nil {
		log.Fatal(err)
		return
	}
	lbp := NewLbp()

	feats := lbp.Process(im, 20)

	fmt.Println(feats[0])
}
