package paint

/*
Canvas implementation using github.com/fogleman/gg
*/
import (
	"image"
	"image/color"
	"image/draw"

	"github.com/fogleman/gg"
)

type ggCanvas struct {
	name string
	ctx  *gg.Context
}

func newFoglemanCanvas(name string, width, height int) *ggCanvas {
	return &ggCanvas{
		name: name,
		ctx:  gg.NewContext(width, height),
	}
}

func newFoglemanCanvasForImage(name string, im image.Image) *ggCanvas {
	return &ggCanvas{
		name: name,
		ctx:  gg.NewContextForImage(im),
	}
}

func (c *ggCanvas) Name() string {
	return c.name
}

func (c *ggCanvas) R(x, y int) uint8 {
	v, _, _, _ := c.At(x, y).RGBA()
	return uint8(v >> 8)
}

func (c *ggCanvas) G(x, y int) uint8 {
	_, v, _, _ := c.At(x, y).RGBA()
	return uint8(v >> 8)
}

func (c *ggCanvas) B(x, y int) uint8 {
	_, _, v, _ := c.At(x, y).RGBA()
	return uint8(v >> 8)
}

func (c *ggCanvas) Image() image.Image {
	return c.ctx.Image()
}

func (c *ggCanvas) SetMask(cnv Canvas) {
	if cnv == nil {
		c.ctx.SetMask(image.NewAlpha(c.Bounds()))
		return
	}
	c.ctx.SetMask(cnv.Mask())
}

func (c *ggCanvas) RectangleHorizontal(r image.Rectangle, colours ...color.Color) {
	if len(colours) == 0 {
		return
	}

	my := float64((r.Max.Y - r.Min.Y) / 2)
	g := gg.NewLinearGradient(float64(r.Min.X), my, float64(r.Max.X), my)

	delta := 1.0 / float64(len(colours))
	for i, c := range colours {
		g.AddColorStop(delta*float64(i), c)
	}

	c.ctx.SetFillStyle(g)
	c.ctx.MoveTo(float64(r.Min.X), float64(r.Min.Y))
	c.ctx.LineTo(float64(r.Max.X), float64(r.Min.Y))
	c.ctx.LineTo(float64(r.Max.X), float64(r.Max.Y))
	c.ctx.LineTo(float64(r.Min.X), float64(r.Max.Y))
	c.ctx.ClosePath()
	c.ctx.Fill()
}

func (c *ggCanvas) RectangleVertical(r image.Rectangle, colours ...color.Color) {
	if len(colours) == 0 {
		return
	}

	mx := float64((r.Max.X - r.Min.X) / 2)
	g := gg.NewLinearGradient(mx, float64(r.Min.Y), mx, float64(r.Max.Y))

	delta := 1.0 / float64(len(colours))
	for i, c := range colours {
		g.AddColorStop(delta*float64(i), c)
	}

	c.ctx.SetFillStyle(g)
	c.ctx.MoveTo(float64(r.Min.X), float64(r.Min.Y))
	c.ctx.LineTo(float64(r.Max.X), float64(r.Min.Y))
	c.ctx.LineTo(float64(r.Max.X), float64(r.Max.Y))
	c.ctx.LineTo(float64(r.Min.X), float64(r.Max.Y))
	c.ctx.ClosePath()
	c.ctx.Fill()
}

func (c *ggCanvas) Line(line []image.Point, width int, colours ...color.Color) {
	if len(line) < 2 {
		return
	}

	a, b := line[0], line[len(line)-1]
	g := gg.NewLinearGradient(float64(a.X), float64(a.Y), float64(b.X), float64(b.Y))

	delta := 1.0 / float64(len(colours))
	for i, c := range colours {
		g.AddColorStop(delta*float64(i), c)
	}

	c.ctx.SetLineWidth(float64(width))
	c.ctx.SetStrokeStyle(g)
	for i := 1; i < len(line); i++ {
		c.ctx.MoveTo(float64(line[i-1].X), float64(line[i-1].Y))
		c.ctx.LineTo(float64(line[i].X), float64(line[i].Y))
	}
	c.ctx.Stroke()
}

func (c *ggCanvas) FlattenOutside(r image.Rectangle) {
	blkctx := gg.NewContext(c.ctx.Width(), c.ctx.Height())
	blkctx.SetColor(color.Black)
	blkctx.DrawRectangle(0, 0, float64(c.ctx.Width()), float64(c.ctx.Height()))
	blkctx.Fill()
	blk := blkctx.Image().(*image.RGBA)

	orig := c.Image()
	draw.Draw(blk, r, orig, r.Min, draw.Src)

	im, _ := SmoothImage(blk, 10)

	ng := im.(*image.NRGBA)
	inset := r.Inset(5)
	draw.Draw(ng, inset, orig, inset.Min, draw.Src)

	c.ctx = gg.NewContextForImage(im)
}

func (c *ggCanvas) Polygon(poly [][2]image.Point, col color.Color) {
	if len(poly) < 3 {
		return
	}

	in := orderedPolygon(poly)

	c.ctx.SetColor(col)
	c.ctx.MoveTo(float64(in[0].X), float64(in[0].Y))
	for i := 1; i < len(in); i++ {
		c.ctx.LineTo(float64(in[i].X), float64(in[i].Y))
	}
	c.ctx.ClosePath()
	c.ctx.Fill()
}

func (c *ggCanvas) Bounds() image.Rectangle {
	return c.ctx.Image().Bounds()
}

func (c *ggCanvas) Smooth(radius uint32) error {
	out, err := smoothImage(c.Image(), radius)
	if err != nil {
		return err
	}
	c.ctx = gg.NewContextForImage(out)
	return nil
}

func (c *ggCanvas) Set(x, y int, col color.Color) {
	c.ctx.SetColor(col)
	c.ctx.SetPixel(x, y)
}

func (c *ggCanvas) At(x, y int) color.Color {
	return c.Image().At(x, y)
}

func (c *ggCanvas) Ellipse(centre image.Point, rx, ry, rot int, depth float64, mode Mode) {
	maj := rx
	if ry > maj {
		maj = ry
	}
	x, y := float64(centre.X), float64(centre.Y)
	c.ctx.Push()
	c.ctx.SetFillStyle(radial(centre, float64(maj)*2, depth, mode))
	c.ctx.RotateAbout(gg.Radians(float64(rot)), x, y)
	c.ctx.DrawEllipse(x, y, float64(rx), float64(ry))
	c.ctx.Fill()
	c.ctx.Pop()
}

func (c *ggCanvas) Channel(path []image.Point, width int, depth float64, mode Mode) {
	if len(path) < 2 {
		return
	}

	fullWidth := float64(width)
	currentWidth := fullWidth
	stops := linear(depth, mode)

	// TODO: it should be possible to make a gg.Gradient that does this, but
	// for now .. this is the inefficient fix ..
	for {
		if currentWidth <= 0 {
			break
		}

		c.ctx.SetColor(getColor(currentWidth/fullWidth, stops))
		c.ctx.SetLineWidth(currentWidth)
		c.ctx.MoveTo(float64(path[0].X), float64(path[0].Y))

		for i := 1; i < len(path); i++ {
			c.ctx.LineTo(float64(path[i].X), float64(path[i].Y))
			c.ctx.Stroke()
			c.ctx.MoveTo(float64(path[i].X), float64(path[i].Y))
		}

		currentWidth -= 1
	}
}
