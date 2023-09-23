package main

import (
	"fmt"
	"github.com/fogleman/gg"
	"image"
	"image/color"
)

type Mode string

const (
	Concave Mode = "concave"
	Convex  Mode = "convex"
)

func main() {
	dc := gg.NewContext(1000, 1000)
	Ellipse(dc, image.Pt(500, 500), 400, 300, 45, Concave)
	dc.SavePNG("out.png")
}

func Ellipse(ctx *gg.Context, centre image.Point, rx, ry, rot int, mode Mode) {
	x, y := float64(centre.X), float64(centre.Y)
	maj := rx
	if rx < ry {
		maj = ry
	}
	ctx.Push()
	ctx.SetFillStyle(radial(centre, maj, mode))
	ctx.RotateAbout(gg.Radians(float64(rot)), x, y)
	ctx.DrawEllipse(x, y, float64(rx), float64(ry))
	fmt.Println("DRAW ellipse", x, y, rx, ry, rot)
	ctx.Fill()
	ctx.Pop()
}

func radial(centre image.Point, r int, mode Mode) gg.Gradient {
	g := gg.NewRadialGradient(
		float64(centre.X),
		float64(centre.Y),
		float64(2),
		float64(centre.X),
		float64(centre.Y),
		float64(r),
	)
	switch mode {
	case Convex:
		g.AddColorStop(0, color.RGBA{255, 255, 255, 255})
		g.AddColorStop(0.3, color.RGBA{120, 120, 120, 255})
		g.AddColorStop(0.5, color.RGBA{80, 80, 80, 255})
		g.AddColorStop(0.75, color.RGBA{50, 50, 50, 255})
		g.AddColorStop(0.95, color.RGBA{20, 20, 20, 255})
		g.AddColorStop(1, color.RGBA{0, 0, 0, 255})
	case Concave:
		g.AddColorStop(0, color.RGBA{0, 0, 0, 255})
		g.AddColorStop(0.1, color.RGBA{5, 5, 5, 255})
		g.AddColorStop(0.25, color.RGBA{20, 20, 20, 255})
		g.AddColorStop(0.5, color.RGBA{50, 50, 50, 255})
		g.AddColorStop(0.75, color.RGBA{100, 100, 100, 255})
		g.AddColorStop(1, color.RGBA{255, 255, 255, 255})
	}
	return g
}
