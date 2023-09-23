package paint

import (
	"image"
	"image/color"
	"path/filepath"

	"github.com/voidshard/mimage"

	"github.com/fogleman/gg"
)

const (
	defaultRoutines = 4
)

type mimCanvas struct {
	name string
	im   *mimage.Mimage
}

func newMimageCanvas(key string, w, h int) (*mimCanvas, error) {
	im, err := mimage.New(
		image.Rect(0, 0, w, h),
		mimage.Directory(key),
		mimage.OperationRoutines(defaultRoutines),
	)
	return &mimCanvas{
		name: filepath.Dirname(key),
		im:   im,
	}, err
}

func loadMimageCanvas(key string) (*mimCanvas, error) {
	im, err := mimage.Load(key)
	return &mimCanvas{
		name: filepath.Dirname(key),
		im:   im,
	}, err
}

func (m *mimCanvas) Set(x, y int, c color.Color) error {
	op := m.im.Draw()
	op.SetColor(c)
	op.SetPixel(x, y)
	return op.Do()
}

func (m *mimCanvas) At(x, y int) (color.Color, error) {
	return m.im.AtOk(x, y)
}

func (m *mimCanvas) R(x, y int) (uint8, error) {
	c, err := m.AtOk(x, y)
	if err != nil {
		return 0, err
	}
	v, _, _, _ := c.RGBA()
	return uint8(v >> 8), nil
}

func (m *mimCanvas) G(x, y int) (uint8, error) {
	c, err := m.AtOk(x, y)
	if err != nil {
		return 0, err
	}
	_, v, _, _ := c.RGBA()
	return uint8(v >> 8), nil
}

func (m *mimCanvas) B(x, y int) (uint8, error) {
	c, err := m.AtOk(x, y)
	if err != nil {
		return 0, err
	}
	_, _, v, _ := c.RGBA()
	return uint8(v >> 8), nil
}

func (m *mimCanvas) SetMask(in Canvas) error {
	im, ok := in.(*mimCanvas)
	if !ok {
		return fmt.Errorf("unsupported canvas %v", in)
	}
	m.SetMask(im)
	return nil
}

func (m *mimCanvas) Flatten(r image.Rectangle) error {
	op := m.im.Draw()
	op.SetColor(color.Black)
	op.DrawRectangle(r)
	op.Fill()
	return op.Do()
}

func (m *mimCanvas) RectangleHorizontal(r image.Rectangle, colours ...color.Color) error {
	if len(colours) == 0 {
		return nil
	}

	my := float64((r.Max.Y - r.Min.Y) / 2)
	g := gg.NewLinearGradient(float64(r.Min.X), my, float64(r.Max.X), my)

	delta := 1.0 / float64(len(colours))
	for i, c := range colours {
		g.AddColorStop(delta*float64(i), c)
	}

	op := m.im.Draw()
	op.SetFillStyle(g)
	op.MoveTo(float64(r.Min.X), float64(r.Min.Y))
	op.LineTo(float64(r.Max.X), float64(r.Min.Y))
	op.LineTo(float64(r.Max.X), float64(r.Max.Y))
	op.LineTo(float64(r.Min.X), float64(r.Max.Y))
	op.ClosePath()
	op.Fill()

	return op.Do()
}

func (m *mimCanvas) RectangleVertical(r image.Rectangle, colours ...color.Color) error {
	if len(colours) == 0 {
		return
	}

	mx := float64((r.Max.X - r.Min.X) / 2)
	g := gg.NewLinearGradient(mx, float64(r.Min.Y), mx, float64(r.Max.Y))

	delta := 1.0 / float64(len(colours))
	for i, c := range colours {
		g.AddColorStop(delta*float64(i), c)
	}

	op.SetFillStyle(g)
	op.MoveTo(float64(r.Min.X), float64(r.Min.Y))
	op.LineTo(float64(r.Max.X), float64(r.Min.Y))
	op.LineTo(float64(r.Max.X), float64(r.Max.Y))
	op.LineTo(float64(r.Min.X), float64(r.Max.Y))
	op.ClosePath()
	op.Fill()

	return op.Do()
}

func (m *mimCanvas) Line(line []image.Point, width int, colours ...color.Color) error {
	if len(line) < 2 {
		return nil
	}

	a, b := line[0], line[len(line)-1]
	g := gg.NewLinearGradient(float64(a.X), float64(a.Y), float64(b.X), float64(b.Y))

	delta := 1.0 / float64(len(colours))
	for i, c := range colours {
		g.AddColorStop(delta*float64(i), c)
	}

	op := m.im.Draw()
	op.SetLineWidth(float64(width))
	op.SetStrokeStyle(g)
	for i := 1; i < len(line); i++ {
		op.MoveTo(float64(line[i-1].X), float64(line[i-1].Y))
		op.LineTo(float64(line[i].X), float64(line[i].Y))
	}
	op.Stroke()

	return op.Do()
}

// Polygon solid colour
func (m *mimCanvas) Polygon(poly [][2]image.Point, c color.Color) error {
	if len(poly) < 3 {
		return nil
	}

	in := orderedPolygon(poly)

	op := m.im.Draw()
	op.SetColor(col)
	op.MoveTo(float64(in[0].X), float64(in[0].Y))
	for i := 1; i < len(in); i++ {
		op.LineTo(float64(in[i].X), float64(in[i].Y))
	}
	op.ClosePath()
	op.Fill()

	return op.Do()
}

// Ellipse with a gradient
func (m *mimCanvas) Ellipse(centre image.Point, rx, ry, rot int, depth float64, mode Mode) error {
	maj := rx
	if ry > maj {
		maj = ry
	}
	x, y := float64(centre.X), float64(centre.Y)

	angle := gg.Radians(float64(rot))

	op := m.im.Draw()

	op.SetFillStyle(radial(centre, float64(maj)*2, depth, mode))
	op.RotateAbout(angle, x, y)
	op.DrawEllipse(x, y, float64(rx), float64(ry))
	op.Fill()
	op.RotateAbout(-1*angle, x, y)
	op.SetFillStyle(nil)

	return op.Do()
}

// Channel is a line with a gradient across it's width (rather than down it's length)
func (m *mimCanvas) Channel(path []image.Point, width int, depth float64, mode Mode) error {
	if len(path) < 2 {
		return nil
	}

	fullWidth := float64(width)
	currentWidth := fullWidth
	stops := linear(depth, mode)

	// TODO: it should be possible to make a gg.Gradient that does this, but
	// for now .. this is the inefficient fix ..

	op := m.im.Draw()
	for {
		if currentWidth <= 0 {
			return op.Do()
		}

		op.SetColor(getColor(currentWidth/fullWidth, stops))
		op.SetLineWidth(currentWidth)

		for i := 1; i < len(path); i++ {
			op.MoveTo(float64(path[i-1].X), float64(path[i-1].Y))
			op.LineTo(float64(path[i].X), float64(path[i].Y))
			op.Stroke()
		}

		currentWidth -= 1
	}
}

func (m *mimCanvas) Smooth(radius uint32) error {
	return nil
}

func (m *mimCanvas) Name() string {
	return m.name
}

func (m *mimCanvas) Bounds() image.Rectangle {
	return m.im.Bounds()
}
