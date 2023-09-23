package voronoi

import (
	"encoding/json"
	"image"
	"math/rand"

	"github.com/voidshard/genesis/internal/dijkstra"
	"github.com/voidshard/voronoi"
)

// graph wraps logic around
//  voronoi diagrams to calculate verticies, edges and build a graph
//  dijkstra's algo to traverse the graph + extra logic around weights along edges
//  tags - user defined sets of points
// In to one struct.
type graph struct {
	// nb. this whole struct is internal and fields are exported for JSON marshal / unmarshal
	// I don't really intend this to be used externally ever

	GraphName string `json:"name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Seed      int64  `json:"seed"`

	DefaultWeight int      `json:"default_weight"`
	WeightNames   []string `json:"weight_names"`

	SiteCentres []image.Point    `json:"sites"`    // centres of cells
	Vertices    []image.Point    `json:"vertices"` // corners of cells
	Edges       [][2]image.Point `json:"edges"`    // edges of cells

	Tags map[string][]image.Point `json:"tags"` // user defined

	Weights map[string][]int `json:"weights"` // tag -> vertex index -> weight
	dij     *dijkstra.Graph

	voro *voronoi.Voronoi
}

// newGraph creates a voronoi diagram
func newGraph(name string, weightNames []string, width, height, defweight, points int, seed int64) (*graph, error) {
	voro, sites, verts, edges, err := randomVoronoi(seed, width, height, points)
	if err != nil {
		return nil, err
	}

	dij, err := dijkstra.New(defweight, weightNames, verts, edges)
	if err != nil {
		return nil, err
	}

	return &graph{
		GraphName:   name,
		Width:       width,
		Height:      height,
		Seed:        seed,
		WeightNames: weightNames,
		SiteCentres: sites,
		Vertices:    verts,
		Edges:       edges,
		dij:         dij,
		Tags:        map[string][]image.Point{},
		voro:        voro,
	}, nil
}

func (g *graph) Points() []image.Point { return g.Vertices }

func (g *graph) Sites() []image.Point { return g.SiteCentres }

func (g *graph) RandomCell() *Cell {
	c := g.voro.SiteByID(rand.Intn(len(g.SiteCentres)))
	return &Cell{parent: c, Site: image.Pt(c.X(), c.Y())}
}

func (g *graph) NeighbouringCells(in []*Cell) ([]*Cell, error) {
	given := map[int]bool{}
	for _, c := range in {
		given[c.parent.ID()] = true
	}

	found := map[int]*voronoi.Neighbour{}
	for _, c := range in {
		for _, neighbour := range c.parent.Neighbours() {
			_, wasGiven := given[neighbour.Site.ID()]
			if wasGiven {
				continue
			}
			found[neighbour.Site.ID()] = neighbour
		}
	}

	res := make([]*Cell, len(found))
	i := 0
	for _, v := range found {
		res[i] = &Cell{parent: v.Site, Site: image.Pt(v.Site.X(), v.Site.Y())}
		i += 1
	}

	return res, nil
}

func (g *graph) NeighbouringPoints(in []image.Point) ([]image.Point, error) {
	return g.dij.Neighbours(in)
}

func (g *graph) Tag(name string, in []image.Point) {
	g.Tags[name] = in
}

func (g *graph) FromTag(name string) ([]image.Point, bool) {
	result, ok := g.Tags[name]
	return result, ok
}

func (g *graph) RandomPoint() image.Point { return g.dij.RandomPoint() }

func (g *graph) ClosestPoint(in image.Point) image.Point {
	// TODO make more efficient this is .. pretty expensive
	dist := pythagoras(in, g.Vertices[0])
	pnt := g.Vertices[0]

	for _, p := range g.Vertices[1:] {
		d := pythagoras(in, p)
		if d < dist {
			dist = d
			pnt = p
		}
	}

	return pnt
}

func (g *graph) IncrWeights(pts []image.Point, delta map[string]int) error {
	return g.dij.IncrWeights(pts, delta)
}

func (g *graph) IncrWeightsOutside(area image.Rectangle, delta map[string]int) error {
	return g.dij.IncrWeightsOutside(area, delta)
}

func (g *graph) Shortest(weight string, a, b image.Point) ([]image.Point, error) {
	return g.dij.Shortest(weight, a, b)
}

func (g *graph) Name() string { return g.GraphName }

func (g *graph) Marshal() ([]byte, error) {
	g.Weights = g.dij.Weights()
	return json.Marshal(g)
}

func (g *graph) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, g)
	if err != nil {
		return err
	}

	dij, err := dijkstra.New(g.DefaultWeight, g.WeightNames, g.Vertices, g.Edges)
	if err != nil {
		return err
	}

	g.dij = dij

	err = g.dij.SetWeights(g.Weights)
	if err != nil {
		return err
	}

	voro, err := rebuildVoronoi(g.Width, g.Height, g.SiteCentres)

	g.voro = voro
	return err
}
