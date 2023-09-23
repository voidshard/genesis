package paint

import (
	"image"
	"image/color"
)

//
func SmoothImage(in image.Image, radius uint32) (image.Image, error) {
	return smoothImage(in, radius)
}

//
func orderedPolygon(in [][2]image.Point) []image.Point {
	to := map[image.Point]image.Point{}
	for _, line := range in {
		to[line[0]] = line[1]
	}

	start := in[0][0]
	ret := []image.Point{start}

	next := in[0][1]
	for {
		if next == start {
			break // we've reached the end
		}

		ret = append(ret, next)
		next = to[next]
	}

	return ret
}

// merge does a simple in-elegant weighted merge
// I suspect there are more efficient ways of doing this .. using draw / masks maybe?
func merge(pnt Painter, area image.Rectangle, weights map[Canvas]float64) (image.Image, error) {
	im := image.NewGray(area)

	canvases := map[image.Image]float64{}
	for cnv, w := range weights {
		if w == 0 {
			continue
		}
		canvases[cnv.Image()] = w
	}
	if len(canvases) == 0 {
		return im, nil
	}

	for x := area.Min.X; x < area.Max.X; x++ {
		for y := area.Min.Y; y < area.Max.Y; y++ {
			v := 0.0
			for cnv, w := range canvases {
				r, _, _, _ := cnv.At(x, y).RGBA()
				v += w * float64(r>>8)
			}
			if v < 0 { // clamp
				v = 0
			} else if v > 255 {
				v = 255
			}
			im.SetGray(x, y, color.Gray{Y: uint8(v)})
		}
	}

	return im, nil
}
