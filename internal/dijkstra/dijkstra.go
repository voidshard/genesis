package dijkstra

import (
	"fmt"
	"image"
	"math/rand"

	"github.com/RyanCarrier/dijkstra"
)

var (
	ErrPointNotFound = fmt.Errorf("failed to find matching vertex for point")
	ErrInvalidTag    = fmt.Errorf("invalid tag - all tags must be given on creation")
)

// graph specifically breaks out the dijkstra / weights part of the fun.
//
// Since our dijkstra impl only allows one set of wieghts we actually have
// a set of internal dijkstra graphs, each with the weights particular to it.
// Ie. the same graph but whose weights differ.
type Graph struct {
	verts      []image.Point
	neighbours [][]int          // vert index -> list of neighbouring verts
	weights    map[string][]int // tag -> vert index -> weight

	pointLookup map[image.Point]int
	dij         map[string]*dijkstra.Graph
}

//
func (g *Graph) Weights() map[string][]int {
	return g.weights
}

//
func (g *Graph) SetWeights(in map[string][]int) error {
	if in == nil {
		return nil
	}
	weights := map[string][]int{}
	for tag, given := range in {
		dij, ok := g.dij[tag]
		if !ok {
			return fmt.Errorf("%w given tag %s", ErrInvalidTag, tag)
		}

		for pid, w := range given {
			if w < 0 {
				w = 0
			}
			neighbours := g.neighbours[pid]
			for _, nid := range neighbours {
				dij.AddArc(nid, pid, int64(w)) // overwrites
			}
		}

		weights[tag] = given
	}
	g.weights = weights
	return nil
}

// New makes a new graph with no weights.
// Since we initialise internal graphs, we need the names of the graphs (weights) up front.
func New(defaultWeight int, weightNames []string, verts []image.Point, edges [][2]image.Point) (*Graph, error) {
	dij := map[string]*dijkstra.Graph{}
	weights := map[string][]int{}
	for _, tag := range weightNames {
		dij[tag] = dijkstra.NewGraph()
		weights[tag] = make([]int, len(verts))
	}

	pl := map[image.Point]int{}
	for i, p := range verts {
		pl[p] = i
		for _, g := range dij {
			g.AddVertex(i)
		}
	}

	ns := make([][]int, len(verts))
	for _, e := range edges {
		id0, ok := pl[e[0]]
		if !ok {
			return nil, fmt.Errorf("%w %v", ErrPointNotFound, e[0])
		}
		id1, ok := pl[e[1]]
		if !ok {
			return nil, fmt.Errorf("%w %v", ErrPointNotFound, e[1])
		}

		neighbours := ns[id0]
		if neighbours == nil {
			neighbours = []int{id1}
		} else {
			neighbours = append(neighbours, id1)
		}
		ns[id0] = neighbours

		neighbours = ns[id1]
		if neighbours == nil {
			neighbours = []int{id0}
		} else {
			neighbours = append(neighbours, id0)
		}
		ns[id1] = neighbours

		for tag, g := range dij {
			values, _ := weights[tag]
			values[id0] = defaultWeight
			values[id1] = defaultWeight
			g.AddArc(id0, id1, int64(defaultWeight))
			g.AddArc(id1, id0, int64(defaultWeight))
		}
	}

	return &Graph{
		verts:       verts,
		pointLookup: pl,
		neighbours:  ns,
		weights:     weights,
		dij:         dij,
	}, nil
}

func (g *Graph) RandomPoint() image.Point {
	return g.verts[rand.Intn(len(g.verts))]
}

func (g *Graph) IncrWeightsOutside(area image.Rectangle, delta map[string]int) error {
	pts := []image.Point{}
	for _, s := range g.verts {
		if s.X <= area.Min.X || s.X >= area.Max.X || s.Y <= area.Min.Y || s.Y >= area.Max.Y {
			pts = append(pts, s)
		}
	}
	return g.IncrWeights(pts, delta)
}

func (g *Graph) IncrWeights(pts []image.Point, delta map[string]int) error {
	for _, p := range pts {
		pid, ok := g.pointLookup[p]
		if !ok {
			return fmt.Errorf("%w %v", ErrPointNotFound, p)
		}

		neighbours := g.neighbours[pid]

		for tag, dw := range delta {
			dij, ok := g.dij[tag]
			if !ok {
				return fmt.Errorf("%w given tag %s", ErrInvalidTag, tag)
			}
			weights, _ := g.weights[tag]
			weights[pid] = weights[pid] + dw
			if weights[pid] < 0 {
				weights[pid] = 0
			}

			for _, nid := range neighbours {
				dij.AddArc(nid, pid, int64(weights[pid])) // overwrites
			}
		}
	}
	return nil
}

func (g *Graph) Neighbours(in []image.Point) ([]image.Point, error) {
	given := map[int]bool{}
	for _, p := range in {
		pid, ok := g.pointLookup[p]
		if !ok {
			return nil, fmt.Errorf("%w %v", ErrPointNotFound, p)
		}
		given[pid] = true
	}

	found := map[image.Point]bool{}
	for pid := range given {
		neighbours := g.neighbours[pid]
		for _, nid := range neighbours {
			_, wasGiven := given[nid]
			if wasGiven {
				continue
			}
			found[g.verts[nid]] = true
		}
	}

	result := []image.Point{}
	for p := range found {
		result = append(result, p)
	}

	return result, nil
}

func (g *Graph) Shortest(tag string, a, b image.Point) ([]image.Point, error) {
	ai, ok := g.pointLookup[a]
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrPointNotFound, a)
	}
	bi, ok := g.pointLookup[b]
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrPointNotFound, b)
	}

	dij, ok := g.dij[tag]
	if !ok {
		return nil, fmt.Errorf("%w given tag %s", ErrInvalidTag, tag)
	}

	best, err := dij.Shortest(ai, bi)
	if err != nil {
		return nil, err
	}
	path := make([]image.Point, len(best.Path))
	for i, id := range best.Path {
		path[i] = g.verts[id]
	}
	return path, nil
}
