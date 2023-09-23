package geography

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/internal/voronoi"
	"github.com/voidshard/genesis/pkg/types"
)

type graphOperation struct {
	p   *types.Project
	pnt paint.Painter

	voro  voronoi.Voronoi
	graph voronoi.Graph

	from    image.Point
	to      image.Point
	maxDist float64
}

func (e *Editor) newGraphOp(proj string, s *types.PathSpec) (*graphOperation, error) {
	p, err := e.project(proj)
	if err != nil {
		return nil, err
	}

	voro := voronoi.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)
	graph, err := e.cachedGraph(voro, p.VoronoiDiagram())
	if err != nil {
		return nil, err
	}

	if s == nil {
		s = &types.PathSpec{MaxDist: float64(p.WorldWidth+p.WorldHeight) / 2 / 3}
	}

	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)

	// use given from / to or pick random points not too close to the edge
	pointA := graph.RandomPoint()
	if s.From != nil {
		pointA = graph.ClosestPoint(*s.From)
	}
	pointB := graph.RandomPoint()
	if s.To != nil {
		pointB = graph.ClosestPoint(*s.To)
	}

	return &graphOperation{
		p:       p,
		voro:    voro,
		graph:   graph,
		pnt:     pnt,
		from:    pointA,
		to:      pointB,
		maxDist: s.MaxDist,
	}, nil
}

func (e *Editor) newVoronoiNoise(p *types.Project, pnt paint.Painter, voro voronoi.Voronoi, points int) (voronoi.Graph, paint.Canvas, error) {
	seed := rand.Int63()

	// build voronoi diagram
	diag, err := voro.NewGraph(
		p.VoronoiDiagram(),
		voroWeights,
		e.set.GraphDefaultWeight,
		points,
		seed,
	)
	if err != nil {
		return nil, nil, err
	}

	// build voronoi noise (fractal noise)
	highPoints := map[int]*voronoi.Cell{}
	cells := map[int]*voronoi.Cell{}

	for i := 0; i < int(float64(len(diag.Sites()))*e.set.NoiseFractalSegments); i++ {
		// pick random cells to elevate
		c := diag.RandomCell()
		highPoints[c.ID()] = c
		cells[c.ID()] = c
	}

	// mark cells & neighbouring cells for elevation
	cellDeltas := map[int]int{}
	for _, middle := range highPoints {
		next := []*voronoi.Cell{middle}
		for i := 0; i < e.set.NoiseFractalIterations; i++ {
			ns, err := diag.NeighbouringCells(next)
			if err != nil {
				return nil, nil, err
			}
			delta := rand.Intn(10) + 5
			for _, c := range append(next, ns...) {
				d, _ := cellDeltas[c.ID()]
				cellDeltas[c.ID()] = d + delta
				cells[c.ID()] = c
			}
			next = append(next, ns...)
		}
	}

	// color cells with their values
	vnoise, err := pnt.Canvas(p.Canvas(tagVoro))
	if err != nil {
		return nil, nil, err
	}
	for id, delta := range cellDeltas {
		cell, ok := cells[id]
		if !ok { // ??
			continue
		}
		d8 := uint8(delta)
		vnoise.Polygon(cell.Edges(), color.RGBA{d8, d8, d8, 255})
	}

	// and smooth so it's not quite so blocky
	err = vnoise.Smooth(6)
	if err != nil {
		return nil, nil, err
	}

	err = pnt.Save(vnoise)

	return diag, vnoise, err
}
