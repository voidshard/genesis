package paint

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/voidshard/mimage"
)

type fsPaint struct {
	root   string
	width  int
	height int
}

func New(root string, width, height int) Painter {
	os.MkdirAll(root, 0750)
	return &fsPaint{
		root:   root,
		width:  width,
		height: height,
	}
}

// NewCanvas returns a blank canvas
func (p *fsPaint) NewCanvas(name string) (Canvas, error) {
	return newMimageCanvas(p.pathFor(name), p.width, p.height), nil
}

// NewPerlinCanvas returns a canvas with some perlin noise on it
func (p *fsPaint) NewPerlinCanvas(name string, scale float64) (Canvas, error) {
	cnv, err := newMimageCanvas(p.pathFor(name), p.width, p.height)
	if err != nil {
		return nil, err
	}

	size := 500

	for x := 0; x < p.width; x += size {
		for y := 0; y < p.height; y += size {
			op := cnv.im.Draw()
			im.DrawImage(NewPerlin(size, size, scale, false), x, y)
			err = op.Do()
			if err != nil {
				return nil, err
			}
		}
	}

	return cnv, nil
}

// Merge canvases together & output the resulting image
func (p *fsPaint) Merge(area image.Rectangle, weights map[Canvas]float64) (image.Image, error) {
	return merge(p, area, weights)
}

// Save given canvas
func (p *fsPaint) Save(in Canvas) error {
	im, ok := in.(*mimCanvas)
	if ok {
		return im.Flush()
	}
	return fmt.Errorf("", in)
}

// Canvas returns canvas if it exists or makes a new one
func (p *fsPaint) Canvas(name string) (Canvas, error) {
	key := p.pathFor(name)

	_, err := os.Stat(key)
	if errors.Is(err, os.ErrNotExist) {
		return p.NewCanvas(name)
	}

	return loadMimageCanvas(key)
}

// Delete existing canvas (noop if it doesn't exist)
func (p *fsPaint) Delete(name string) error {
	return os.RemoveAll(p.pathFor(name))
}

// pathFor retrns where on the disk we store this named graph
func (p *fsPaint) pathFor(name string) string {
	return filepath.Join(p.root, name)
}
