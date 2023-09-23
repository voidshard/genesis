package geography

import (
	"image"
	"sort"
	"sync"
)

// weightEdgeFn determines the weight value of a proposed edge,
// (used to enforce maxExternal)
//
// Smaller is better (down to 0).
// Less than 0 (ie. -1) is considered "edge invalid"
type weightEdgeFn func(a, b int) float64

// Grid bucket divides an area into x-y segements "buckets" and
// recalls which point(s) are in which
type Grid struct {
	width  int
	height int
	size   int
	wrap   bool

	gridWidth  int
	gridHeight int

	neighbours [][]int
}

type Edge struct {
	A, B   int
	Weight float64
}

func NewGridBucket(w, h, gridsize int, wrap bool) *Grid {
	gx := w / gridsize
	gy := h / gridsize

	return &Grid{
		width:      w,
		height:     h,
		size:       gridsize,
		wrap:       wrap,
		gridWidth:  gx,
		gridHeight: gy,
		neighbours: make([][]int, gx*gy),
	}
}

// index finds the correct numbered bucket for a point.
// nb. we lose information doing this (because of integer division)
// so we can't go backwards.
func (g *Grid) index(p image.Point) int {
	// int index = y * w + x;
	gx := p.X / g.size
	gy := p.Y / g.size
	return gy*g.gridWidth + gx
}

func (g *Grid) indexLeftRight(i int) (int, int) {
	if i == -1 {
		return -1, -1
	}

	remainderWidth := i % g.gridWidth

	left := i - 1
	if remainderWidth == 0 { // far left
		if g.wrap {
			left = i + g.gridWidth - 1 // wrap to right edge
		} else {
			left = -1
		}
	}

	right := i + 1
	if remainderWidth == g.gridWidth-1 { // far right
		if g.wrap {
			right = i - remainderWidth // wrap to left edge
		} else {
			right = -1
		}
	}

	return left, right
}

func (g *Grid) indexAboveBelow(i int) (int, int) {
	if i == -1 {
		return -1, -1
	}

	above := i - g.gridWidth
	if above < 0 {
		if g.wrap {
			above = g.gridWidth*(g.gridHeight-1) + i
		} else {
			above = -1
		}
	}

	below := i + g.gridWidth
	if below >= g.gridWidth*g.gridHeight {
		if g.wrap {
			below = g.gridWidth * (g.gridHeight - 1)
		} else {
			below = -1
		}
	}

	return above, below
}

//
func (g *Grid) BuildEdges(intEdge, extEdge weightEdgeFn, maxExternal int) <-chan [2]int {
	rchan := make(chan [2]int)
	wg := &sync.WaitGroup{}
	for n := 0; n < len(g.neighbours); n++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// consider edges between points in this grid
			ns := g.neighbours[index]
			if ns == nil || len(ns) < 2 {
				return
			}

			for i := 0; i < len(ns)-1; i++ {
				for j := i + 1; j < len(ns)-1; j++ {
					if intEdge(ns[i], ns[j]) >= 0 {
						rchan <- [2]int{ns[i], ns[j]}
					}
				}
			}

			// consider edges between points in this grid
			// and points in neighbouring grids
			for _, neighbours := range g.neighbouringGrids(index) {
				edges := []*Edge{}
				for i := 0; i < len(ns)-1; i++ {
					for j := 0; j < len(neighbours)-1; j++ {
						a, b := ns[i], neighbours[j]
						w := extEdge(a, b)
						if w < 0 {
							continue
						}
						edges = append(edges, &Edge{A: a, B: b, Weight: w})
					}
				}
				if len(edges) >= maxExternal { // need to sort
					sort.Slice(edges, func(i, j int) bool {
						return edges[i].Weight < edges[j].Weight
					})
				}
				for i := 0; i < minInt(maxExternal, len(edges)); i++ {
					rchan <- [2]int{edges[i].A, edges[i].B}
				}
			}
		}(n)
	}

	go func() {
		// ensure the channel is closed so the caller knows we're done
		wg.Wait()
		close(rchan)
	}()

	return rchan
}

func (g *Grid) neighbouringGrids(i int) [][]int {
	left, right := g.indexLeftRight(i)
	above, below := g.indexAboveBelow(i)
	al, ar := g.indexLeftRight(above)
	bl, br := g.indexLeftRight(below)

	indexes := []int{left, right, above, below, al, ar, bl, br}
	res := [][]int{}

	for _, n := range indexes {
		if n == -1 {
			continue
		}
		ns := g.neighbours[n]
		if ns == nil || len(ns) == 0 {
			continue
		}
		res = append(res, ns)
	}

	return res
}

// NeighbouringGrids returns all points from grid sections immediately adjoining
// a grid with the given point.
// Nb. this does *not* return points in the *same* grid as the given point
// (because we have NeighbouringPoints for that).
func (g *Grid) NeighbouringGrids(p image.Point) [][]int {
	return g.neighbouringGrids(g.index(p))
}

// NeighbouringPoints returns points within the same grid as the given point.
// Nb. this doesn't consider that points in a neighbouring grid might
// actually be closer.
func (g *Grid) NeighbouringPoints(p image.Point) []int {
	i := g.index(p)
	ns := g.neighbours[i]
	if ns == nil { // sugar to not return a nil
		return []int{}
	}
	return ns
}

// AddPoint .. adds a point
func (g *Grid) AddPoint(num int, p image.Point) {
	i := g.index(p)
	ns := g.neighbours[i]
	if ns == nil {
		ns = []int{num}
	} else {
		ns = append(ns, num)
	}
	g.neighbours[i] = ns
}
