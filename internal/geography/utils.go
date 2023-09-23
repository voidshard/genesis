package geography

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/voidshard/genesis/pkg/types"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func edgePoints(r image.Rectangle, h types.Heading) <-chan image.Point {
	ch := make(chan image.Point)

	go func() {
		defer close(ch)

		switch h {
		case types.NORTH:
			for x := r.Min.X; x < r.Max.X; x++ {
				ch <- image.Pt(x, r.Min.Y)
			}
		case types.SOUTH:
			for x := r.Min.X; x < r.Max.X; x++ {
				ch <- image.Pt(x, r.Max.Y)
			}
		case types.WEST:
			for y := r.Min.Y; y < r.Max.Y; y++ {
				ch <- image.Pt(r.Min.X, y)
			}
		case types.EAST:
			for y := r.Min.Y; y < r.Max.Y; y++ {
				ch <- image.Pt(r.Max.X, y)
			}
		case types.NORTHEAST:
			for x := r.Min.X; x < r.Max.X; x++ {
				ch <- image.Pt(x, r.Min.Y)
			}
			for y := r.Min.Y; y < r.Max.Y; y++ {
				ch <- image.Pt(r.Max.X, y)
			}
		case types.SOUTHEAST:
			for x := r.Min.X; x < r.Max.X; x++ {
				ch <- image.Pt(x, r.Max.Y)
			}
			for y := r.Min.Y; y < r.Max.Y; y++ {
				ch <- image.Pt(r.Max.X, y)
			}
		case types.SOUTHWEST:
			for x := r.Min.X; x < r.Max.X; x++ {
				ch <- image.Pt(x, r.Max.Y)
			}
			for y := r.Min.Y; y < r.Max.Y; y++ {
				ch <- image.Pt(r.Min.X, y)
			}
		case types.NORTHWEST:
			for x := r.Min.X; x < r.Max.X; x++ {
				ch <- image.Pt(x, r.Min.Y)
			}
			for y := r.Min.Y; y < r.Max.Y; y++ {
				ch <- image.Pt(r.Min.X, y)
			}
		}
	}()

	return ch
}

func forceUint8(a int) uint8 {
	if a <= 0 {
		return 0
	}
	if a >= 255 {
		return 255
	}
	return uint8(a)
}

func incrUint8(a uint8, in float64) uint8 {
	if in <= 0 {
		return a
	} else if in >= 255 {
		return 255
	}
	b := uint8(in)
	if a+b > 255 {
		return 255
	} else if a+b < 0 {
		return 0
	}
	return a + b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// splitUint16 turns a uint16 into two uint8
func splitUint16(in uint16) (uint8, uint8) {
	return uint8(in >> 8), uint8(in)
}

// combineUint16 turns two uint8 into a uint16
func combineUint16(a, b uint8) uint16 {
	return uint16(int(a)<<8 + int(b))
}

//
func fanIn(errs chan error, wg *sync.WaitGroup) error {
	var err error

	ewg := &sync.WaitGroup{}
	ewg.Add(1)

	go func() {
		// collect errors
		defer ewg.Done()
		for e := range errs {
			if e == nil {
				continue
			}
			if err == nil {
				err = e
			} else {
				err = fmt.Errorf("%w %v", err, e)
			}
		}
	}()

	wg.Wait()   // wait for workers to finish
	close(errs) // signal no more errors are coming
	ewg.Wait()  // wait for err chan to finish
	return err  // return final err
}

// pointNear returns a point nearby to `p`
func pointNear(p image.Point, min int, max int) image.Point {
	dx := rand.Intn(max-min) + min
	dy := rand.Intn(max-min) + min
	if rand.Intn(2) == 1 {
		dx *= -1
	}
	if rand.Intn(2) == 1 {
		dy *= -1
	}
	return image.Pt(p.X+dx, p.Y+dy)
}

// distBetween standard pythag.
func distBetween(ax, ay, bx, by int) float64 {
	return math.Sqrt(math.Pow(float64(ax-bx), 2) + math.Pow(float64(ay-by), 2))
}

//
func trimPath(path []image.Point, max float64) []image.Point {
	total := 0.0
	for i := 1; i < len(path); i++ {
		total += distBetween(path[i-1].X, path[i-1].Y, path[i].X, path[i].Y)
		if total >= max {
			return path[0:i]
		}
	}
	return path
}
