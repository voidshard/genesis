package paint

import (
	"image"
	"image/color"

	"github.com/fogleman/gg"
)

type stop struct {
	pos   float64
	color color.Color
}

func linear(depth float64, mode Mode) []*stop {
	if depth < 0 {
		depth = 0
	} else if depth > 1 {
		depth = 1
	}

	switch mode {
	case Convex:
		return []*stop{
			&stop{0, grey(255, depth)},
			&stop{0.5, grey(128, depth)},
			&stop{1, grey(0, 0)},
		}
	case Concave:
		return []*stop{
			&stop{0, grey(0, 0)},
			&stop{0.5, grey(128, depth)},
			&stop{1, grey(255, depth)},
		}
	}
	return nil
}

func radial(centre image.Point, radius, depth float64, mode Mode) gg.Gradient {
	g := gg.NewRadialGradient(
		float64(centre.X),
		float64(centre.Y),
		float64(2),
		float64(centre.X),
		float64(centre.Y),
		radius,
	)

	if depth < 0 {
		depth = 0
	} else if depth > 1 {
		depth = 1
	}

	switch mode {
	case Convex:
		g.AddColorStop(0, grey(255, depth))
		g.AddColorStop(0.3, grey(128, depth))
		g.AddColorStop(0.5, grey(80, depth))
		g.AddColorStop(0.75, grey(50, depth))
		g.AddColorStop(0.95, grey(25, depth))
		g.AddColorStop(1, grey(0, 0))
	case Concave:
		g.AddColorStop(0, grey(0, 0))
		g.AddColorStop(0.1, grey(5, depth))
		g.AddColorStop(0.25, grey(20, depth))
		g.AddColorStop(0.5, grey(50, depth))
		g.AddColorStop(0.75, grey(100, depth))
		g.AddColorStop(1, grey(255, depth))
	}
	return g
}

func grey(v uint8, weight float64) color.RGBA {
	if weight < 0 {
		return color.RGBA{0, 0, 0, 255}
	}
	v = uint8(weight * float64(v))
	return color.RGBA{v, v, v, 255}
}

func colorLerp(c0, c1 color.Color, t float64) color.Color {
	r0, g0, b0, a0 := c0.RGBA()
	r1, g1, b1, a1 := c1.RGBA()

	return color.RGBA{
		lerp(r0, r1, t),
		lerp(g0, g1, t),
		lerp(b0, b1, t),
		lerp(a0, a1, t),
	}
}

func lerp(a, b uint32, t float64) uint8 {
	return uint8(int32(float64(a)*(1.0-t)+float64(b)*t) >> 8)
}

func getColor(pos float64, stops []*stop) color.Color {
	if pos <= 0.0 || len(stops) == 1 {
		return stops[0].color
	}

	last := stops[len(stops)-1]

	if pos >= last.pos {
		return last.color
	}

	for i, stop := range stops[1:] {
		if pos < stop.pos {
			pos = (pos - stops[i].pos) / (stop.pos - stops[i].pos)
			return colorLerp(stops[i].color, stop.color, pos)
		}
	}

	return last.color
}
