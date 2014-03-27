// This is an implementation of VLFeat's local binary pattern library
// written in Go. It implements the local binary patterns algorithm
// described by (Ojala, Pietikainen, & Harwood, 1996)
package lbp

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path"
	"strings"
	"sync"
)

type feature struct {
	data  []float64
	Class string
}

// The Lbp struct holds information relevant to the extraction
// of local binary patterns.
//
// The dim field stores the dimension (radius) of the LBP  operator
// the mapping field is used to reduce the dimensionality of the
// output of the original LBP operator as per (Maturana, Mery, & Soto, 2009).
// Currently only "uniform patterns" are supported.
type Lbp struct {
	dim     int
	mapping []uint8
}

// Initializer function for creating a new Lbp struct. It returns a pointer
// to the newly created struct and fills in its fields appropriately.
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

// All preprocessing required before running the LBP operator takes place here.
// Currently all this does is opens the image and converts it to grayscale.
func (lbp *Lbp) preprocess(path string) ([]uint8, image.Rectangle, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, image.Rectangle{}, err
	}
	defer r.Close()

	im, _, err := image.Decode(r)
	if err != nil {
		return nil, image.Rectangle{}, err
	}

	bounds := im.Bounds()
	grayIm := make([]uint8, bounds.Max.Y*bounds.Max.X)

	// Store the image as a Y * X uint8 array (grayscale)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(im.At(x, y)).(color.Gray)
			grayIm[x+y*bounds.Max.X] = gray.Y
		}
	}

	return grayIm, bounds, err
}

// The heart of the LBP algorithm. The image is treated as being composed of
// distinct regions each of size regionSize x regionSize. The LBP operator is
// run over the entire image, and a histogram of its (quantized) values is
// stored for each region. These histograms are stored continguously in the
// only return value of this function.
func (lbp *Lbp) Process(path string, regionSize int) (*feature, error) {
	im, bounds, err := lbp.preprocess(path)
	if err != nil {
		return nil, err
	}

	maxX := bounds.Max.X
	maxY := bounds.Max.Y

	// rwidth := maxX / regionSize
	// rheight := maxY / regionSize
	rwidth := 50
	rheight := 50
	rstride := rwidth * rheight

	var rSizeX, rSizeY float32
	rSizeX = float32(maxX) / float32(rwidth)
	rSizeY = float32(maxY) / float32(rheight)

	feat := new(feature)
	feat.Class = path
	feat.data = make([]float64, lbp.dim*rstride)

	var bits uint8
	var wy1, wy2, wx1, wx2 float64
	var ry1, ry2, rx1, rx2 int
	for y := 1; y < maxY-1; y++ {

		wy1 = (float64(y)+0.5)/float64(rSizeY) - 0.5
		ry1 = int(math.Floor(wy1))
		ry2 = ry1 + 1
		wy2 = wy1 - float64(ry1)
		wy1 = 1.0 - wy2

		if ry1 >= rheight {
			continue
		}

		for x := 1; x < maxX-1; x++ {
			wx1 = (float64(x)+0.5)/float64(rSizeX) - 0.5
			rx1 = int(math.Floor(wx1))
			rx2 = rx1 + 1
			wx2 = wx1 - float64(rx1)
			wx1 = 1.0 - wx2

			if rx1 >= rwidth {
				continue
			}

			bits = 0
			center := at(im, x, y, maxX)
			if at(im, x+1, y, maxX) > center {
				bits |= 0x1 << 0 // E
			}
			if at(im, x+1, y+1, maxX) > center {
				bits |= 0x1 << 1 // SE
			}
			if at(im, x+0, y+1, maxX) > center {
				bits |= 0x1 << 2 // S
			}
			if at(im, x-1, y+1, maxX) > center {
				bits |= 0x1 << 3 // SW
			}
			if at(im, x-1, y+0, maxX) > center {
				bits |= 0x1 << 4 // W
			}
			if at(im, x-1, y-1, maxX) > center {
				bits |= 0x1 << 5 // NW
			}
			if at(im, x+0, y-1, maxX) > center {
				bits |= 0x1 << 6 // N
			}
			if at(im, x+1, y-1, maxX) > center {
				bits |= 0x1 << 7 // NW
			}

			bin := int(lbp.mapping[bits])

			if rx1 >= 0 && ry1 >= 0 {
				feat.data[bin*rstride+ry1*rwidth+rx1] += wx1 * wy1
			}
			if rx2 < rwidth && ry1 >= 0 {
				feat.data[bin*rstride+ry1*rwidth+rx2] += wx2 * wy1
			}
			if rx1 >= 0 && ry2 < rheight {
				feat.data[bin*rstride+ry2*rwidth+rx1] += wx1 * wy2
			}
			if rx2 < rwidth && ry2 < rheight {
				feat.data[bin*rstride+ry2*rwidth+rx2] += wx2 * wy2
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
				norm += feat.data[k*rstride+offset]
			}
			norm = math.Sqrt(norm) + 1e-10
			for k := 0; k < lbp.dim; k++ {
				feat.data[k*rstride+offset] = math.Sqrt(feat.data[k*rstride+offset]) / norm
			}
		}
		offset++
	}

	return feat, err
}

func (lbp *Lbp) ProcessAll(dirPath string, nWorkers int) ([]feature, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}

	entries, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	// feats := make([][]float64, 1)
	var feats []feature

	images := make(chan string)
	go func() {
		for _, fname := range entries {
			if !isImage(fname) {
				continue
			}
			images <- path.Join(dirPath, fname)
		}
		close(images)
	}()

	var done sync.WaitGroup
	for i := 0; i < nWorkers; i++ {
		done.Add(1)
		go func() {
			for fpath := range images {
				h, err := lbp.Process(fpath, 20)
				if err != nil {
					fmt.Println(err)
				}
				if err == nil {
					feats = append(feats, *h)
				}
			}
			done.Done()
		}()
	}
	done.Wait()
	return feats, nil

}

// Convenience function for using 2-d indexing on a
// 1-d array.
func at(image []uint8, x int, y int, w int) uint8 {
	return image[x+y*w]
}

func isImage(fname string) bool {
	if !strings.Contains(fname, ".jpg") && !strings.Contains(fname, ".png") && !strings.Contains(fname, ".png") {
		return false
	}
	return true

}
