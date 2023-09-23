package paint

import (
	"image"
	"image/color"
)

type Mode string

const (
	Concave Mode = "concave"
	Convex  Mode = "convex"
)

type Painter interface {
	// NewCanvas returns a blank canvas
	NewCanvas(name string) (Canvas, error)

	// NewPerlinCanvas returns a canvas with some perlin noise on it
	NewPerlinCanvas(name string, noise float64) (Canvas, error)

	// NewCanvasFromImage returns a canvas based on the given image
	NewCanvasFromImage(name string, im image.Image) (Canvas, error)

	// Canvas loads, canvas if it exists, otherwise returns a new blank
	Canvas(name string) (Canvas, error)

	// Delete existing canvas (noop if it doesn't exist)
	Delete(name string) error

	// Save given canvas
	Save(Canvas) error

	// Merge `area` of canvases, weighted into one greyscale image.
	Merge(area image.Rectangle, weights map[Canvas]float64) (image.Image, error)
}

type Canvas interface {
	Set(x, y int, c color.Color) error

	At(x, y int) (color.Color, error)

	R(x, y int) (uint8, error)
	G(x, y int) (uint8, error)
	B(x, y int) (uint8, error)

	SetMask(in Canvas) error

	Flatten(r image.Rectangle) error

	RectangleHorizontal(r image.Rectangle, colours ...color.Color) error

	RectangleVertical(r image.Rectangle, colours ...color.Color) error

	Line(line []image.Point, width int, colours ...color.Color) error

	// Polygon solid colour
	Polygon(outline [][2]image.Point, c color.Color) error

	// Ellipse with a gradient
	Ellipse(centre image.Point, rx, ry, rotationDegrees int, depth float64, mode Mode) error

	// Channel is a line with a gradient across it's width (rather than down it's length(
	Channel(pts []image.Point, width int, depth float64, mode Mode) error

	Smooth(radius uint32) error

	Name() string

	Bounds() image.Rectangle
}
