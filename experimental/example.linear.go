package main

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"github.com/fogleman/gg"
)

func main() {
	//Testing making gradiant that varies over the width of a line
	//rather than it's length
	dc := gg.NewContext(500, 400)

	grad := NewLinearLengthwiseGradient(10, 10, 400, 400, 10)
	grad.AddColorStop(1, color.RGBA{0, 0, 0, 255})
	grad.AddColorStop(0.5, color.RGBA{128, 128, 128, 255})
	grad.AddColorStop(0, color.RGBA{255, 255, 255, 255})

	dc.SetStrokeStyle(grad)
	dc.SetLineWidth(20)
	dc.MoveTo(10, 10)
	dc.LineTo(400, 400)
	dc.Stroke()

	dc.SavePNG("out.png")
}

// Linear Gradient
type linearGradient struct {
	x0, y0, x1, y1 float64
	m              float64 // 'm' or gradient of y = mx + c
	c              float64 // 'c' or constant of y = mx + c
	stops          stops
	width          float64
}

func (g *linearGradient) ColorAt(x, y int) color.Color {
	if len(g.stops) == 0 {
		return color.Transparent
	}

	fx, fy := float64(x), float64(y)
	dx, dy := g.x0-g.x1, g.y0-g.y1
	mag := 0.0

	if dy == 0 { // horizontal
		mag = math.Abs(fy - g.y0)
	} else if dx == 0 { // vertical
		mag = math.Abs(fx - g.x0)
	} else {
		// find m and c of tangent to g passing through fx,fy
		tm := -1 * g.m
		tc := linearConstant(fx, fy, tm)

		// find point where lines intersect
		x0 := lineIntersectionX(g.m, g.c, tm, tc)
		y0 := tm*fx + tc // calculate y from tangent x m and c

		fmt.Printf("line y=%vx + %v\n", g.m, g.c)
		fmt.Printf("tan y=%vx + %v\n", tm, tc)
		fmt.Printf("full: (%v, %v) (%v, %v)\n", x0, y0, fx, fy)
		mag = math.Hypot(fx-x0, fy-y0)
	}

	// Calculate distance to (fx, fy) along (x0,y0)->(x1,y1)
	fmt.Println("mag", mag, "width", g.width, mag/g.width)
	return getColor(mag/g.width, g.stops)
}

func (g *linearGradient) AddColorStop(offset float64, color color.Color) {
	g.stops = append(g.stops, stop{pos: offset, color: color})
	sort.Sort(g.stops)
}

func NewLinearLengthwiseGradient(x0, y0, x1, y1, width float64) gg.Gradient {
	m := (y1 - y0) / (x1 - x0)
	return &linearGradient{
		x0: x0, y0: y0,
		x1: x1, y1: y1,
		m:     m,
		c:     linearConstant(x0, y0, m),
		width: width,
	}
}

func linearConstant(x0, y0, m float64) float64 {
	// c = y - mx from rearranging y = mx + c
	return y0 - m*x0
}

// given the m and c values of two lines (y = mx+c)
// return the x value where the two lines intersect.
func lineIntersectionX(m0, c0, m1, c1 float64) float64 {
	// rearranging m0x + c0 = m1x + c1
	// becomes x = (c1 - c0) / (m0 - m1)
	return (c1 - c0) / (m0 - m1)
}

func getColor(pos float64, stops stops) color.Color {
	if pos <= 0.0 || len(stops) == 1 {
		fmt.Println("get color", pos, 0)
		return stops[0].color
	}

	last := stops[len(stops)-1]

	if pos >= last.pos {
		fmt.Println("get color", pos, len(stops)-1)
		return last.color
	}

	for i, stop := range stops[1:] {
		if pos < stop.pos {
			fmt.Println("get color", pos, i, stop.pos)
			pos = (pos - stops[i].pos) / (stop.pos - stops[i].pos)
			return colorLerp(stops[i].color, stop.color, pos)
		}
	}

	fmt.Println("get color (last)", pos, len(stops)-1)
	return last.color
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

type stop struct {
	pos   float64
	color color.Color
}

type stops []stop

// Len satisfies the Sort interface.
func (s stops) Len() int {
	return len(s)
}

// Less satisfies the Sort interface.
func (s stops) Less(i, j int) bool {
	return s[i].pos < s[j].pos
}

// Swap satisfies the Sort interface.
func (s stops) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
