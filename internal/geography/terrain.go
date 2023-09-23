package geography

import (
	"fmt"
	"image"
	"math/rand"
	"sync"

	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/internal/voronoi"
	"github.com/voidshard/genesis/pkg/types"
)

func (e *Editor) CreateTectonics(proj string, noise float64, points int) error {
	p, err := e.project(proj)
	if err != nil {
		return err
	}

	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)
	voro := voronoi.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)

	errchan := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	var diag voronoi.Graph
	var vnoise paint.Canvas
	var verr error

	go func() {
		defer wg.Done()

		// build voronoi noise (fractal noise)
		diag, vnoise, verr = e.newVoronoiNoise(p, pnt, voro, points)
		errchan <- verr
	}()

	e.graph = diag // cache graph we'll probably be using ..

	var pnoise paint.Canvas
	go func() {
		defer wg.Done()

		// build perlin noise (smooth noise)
		pcnv, err := pnt.NewPerlinCanvas(p.Canvas(tagPerlin), noise)
		if err != nil {
			errchan <- err
		}
		pnoise = pcnv
		err = pnt.Save(pcnv)
		errchan <- err
	}()

	err = fanIn(errchan, wg)
	if err != nil {
		return err
	}

	// weight voronoi based on noise values at vertexes
	perlinImg := pnoise.Image()
	voroImg := vnoise.Image()
	for _, p := range diag.Points() {
		pnValue, _, _, _ := perlinImg.At(p.X, p.Y).RGBA()
		flValue, _, _, _ := voroImg.At(p.X, p.Y).RGBA()

		v := int((float64(pnValue>>8)*e.set.HeightMapNoisePerlinWeight + float64(flValue)*e.set.HeightMapNoiseVoronoiWeight) / 2)
		diag.IncrWeights([]image.Point{p}, map[string]int{
			tagMountains: -1 * v,
			tagVolcanoes: -1 * v / 2,
			tagRavines:   v,
			tagRivers:    v * 2,
		})
	}

	// weight vertexes along edges (encourage stuff to avoid sides)
	deltaOnEdge := map[string]int{}
	for _, w := range voroWeights {
		deltaOnEdge[w] = e.set.GraphEdgeWeight
	}
	bounds := image.Rect(5, 5, p.WorldWidth-5, p.WorldHeight-5)

	err = diag.IncrWeightsOutside(bounds, deltaOnEdge)
	if err != nil {
		return err
	}

	return voro.Save(diag)
}

//
func (e *Editor) FlattenOutside(proj string, bounds image.Rectangle) error {
	p, err := e.project(proj)
	if err != nil {
		return nil
	}

	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)

	errchan := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()
		cnv, err := pnt.Canvas(p.Canvas(tagMountains))
		if err != nil {
			errchan <- err
			return
		}
		cnv.FlattenOutside(bounds)
		errchan <- pnt.Save(cnv)
	}()

	go func() {
		defer wg.Done()
		cnv, err := pnt.Canvas(p.Canvas(tagPerlin))
		if err != nil {
			errchan <- err
			return
		}
		cnv.FlattenOutside(bounds)
		errchan <- pnt.Save(cnv)
	}()

	go func() {
		defer wg.Done()
		cnv, err := pnt.Canvas(p.Canvas(tagVoro))
		if err != nil {
			errchan <- err
			return
		}
		cnv.FlattenOutside(bounds)
		errchan <- pnt.Save(cnv)
	}()

	return fanIn(errchan, wg)
}

//
func (e *Editor) SmoothTerrain(proj string, radius uint32) error {
	p, err := e.project(proj)
	if err != nil {
		return err
	}

	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)
	mountains, err := pnt.Canvas(p.Canvas(tagMountains))
	if err != nil {
		return err
	}

	mountains.Smooth(radius)

	return pnt.Save(mountains)
}

//
func (e *Editor) HeightMap(proj string, area image.Rectangle) (image.Image, error) {
	op, err := e.newGraphOp(proj, nil)
	if err != nil {
		return nil, err
	}

	mountains, err := op.pnt.Canvas(op.p.Canvas(tagMountains))
	if err != nil {
		return nil, err
	}
	pnNoise, err := op.pnt.Canvas(op.p.Canvas(tagPerlin))
	if err != nil {
		return nil, err
	}
	viNoise, err := op.pnt.Canvas(op.p.Canvas(tagVoro))
	if err != nil {
		return nil, err
	}
	ravines, err := op.pnt.Canvas(op.p.Canvas(tagRavines))
	if err != nil {
		return nil, err
	}
	rivers, err := op.pnt.Canvas(op.p.Canvas(tagRivers))
	if err != nil {
		return nil, err
	}

	wfull := map[paint.Canvas]float64{
		mountains: e.set.HeightMapMountainWeight,
		ravines:   e.set.HeightMapRavineWeight,
		rivers:    e.set.HeightMapRiverWeight,
		pnNoise:   e.set.HeightMapNoisePerlinWeight,
		viNoise:   e.set.HeightMapNoiseVoronoiWeight,
	}

	im, err := op.pnt.Merge(area, wfull)
	if err != nil {
		return nil, err
	}

	final, err := paint.SmoothImage(im, 3)
	e.hmap[area] = final // cached for other internal funcs to call
	return final, err
}

//
func (e *Editor) AddRavine(proj, tag string, s *types.PathSpec, forkChance float64) ([]image.Point, error) {
	op, err := e.newGraphOp(proj, s)
	if err != nil {
		return nil, err
	}

	// find path of ravine
	path, err := op.graph.Shortest(tagRavines, op.from, op.to)
	if err != nil {
		return nil, fmt.Errorf("%w %v", ErrNoPath, err)
	}
	path = trimPath(path, op.maxDist)
	if len(path) < 2 {
		return nil, fmt.Errorf("%w path too short", ErrNoPath)
	}

	cnv, err := op.pnt.Canvas(op.p.Canvas(tagRavines))
	if err != nil {
		return nil, err
	}

	// mark ravine on mountain too
	err = op.graph.IncrWeights(
		path,
		map[string]int{
			tagRavines:   e.set.GraphRavineWeight,
			tagMountains: e.set.GraphRavineWeight,
			tagRivers:    e.set.GraphRavineWeight * -1,
			tagVolcanoes: e.set.GraphRavineWeight,
		},
	)
	if err != nil {
		return nil, err
	}

	if tag != "" {
		op.graph.Tag(tag, path)
	}

	// draw the ravine
	cnv.Channel(
		path,
		e.set.RavineWidth.Roll(),
		rand.Float64()/5+0.8,
		paint.Convex, // we'll treat > 0 as "low". Eg this map is inverted
	)

	// save everything and return
	err = op.voro.Save(op.graph)
	if err != nil {
		return nil, err
	}
	return path, op.pnt.Save(cnv)
}

//
func (e *Editor) AddMountainRange(proj, tag string, s *types.PathSpec, scale float64) ([]image.Point, []image.Point, error) {
	op, err := e.newGraphOp(proj, s)
	if err != nil {
		return nil, nil, err
	}

	// find segments on voronoi that link the ends, mark as mountains
	path, err := op.graph.Shortest(tagMountains, op.from, op.to)
	if err != nil {
		return nil, nil, fmt.Errorf("%w %v", ErrNoPath, err)
	}
	path = trimPath(path, op.maxDist)
	if len(path) < 2 {
		return nil, nil, fmt.Errorf("%w path too short", ErrNoPath)
	}

	cnv, err := op.pnt.Canvas(op.p.Canvas(tagMountains))
	if err != nil {
		return nil, nil, err
	}

	errchan := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		// weight mountains themselves
		err = op.graph.IncrWeights(
			path,
			map[string]int{
				tagMountains: e.set.GraphMountainWeight,
				tagRavines:   e.set.GraphMountainWeight,
				tagRivers:    e.set.GraphMountainWeight,
				tagVolcanoes: e.set.GraphMountainWeight * -1,
			},
		)
		if err != nil {
			errchan <- err
			return
		}

		// weight nearby to create mountain passes
		nearby, err := op.graph.NeighbouringPoints(path)
		if err != nil {
			errchan <- err
			return
		}

		err = op.graph.IncrWeights(
			nearby,
			map[string]int{
				tagMountains: e.set.GraphMountainWeight,
				tagRavines:   e.set.GraphMountainWeight * -1,
				tagRivers:    e.set.GraphMountainWeight * -1,
			},
		)
		if err != nil {
			errchan <- err
			return
		}

		// tag mountain range on map
		if tag != "" {
			op.graph.Tag(tag, path)
		}

		err = op.voro.Save(op.graph)
		if err != nil {
			errchan <- err
			return
		}
	}()

	placed := []image.Point{}
	go func() {
		defer wg.Done()

		// draw mountains along the path
		for j := 1; j < len(path); j++ { // for each segment of the range
			// walk along the segment
			alongLine := voronoi.PointsBetween(path[j-1], path[j])
			rangeHeight := 3 * rand.Float64() / 4
			for p := 0; p < len(alongLine); {
				// pick a point along the segment
				centre := alongLine[p]
				for mnt := 0; mnt < e.set.MountainsPerStep.Roll(); mnt++ {
					// place mountain centred around segment point
					cp := image.Pt(
						centre.X-e.set.MountainRangeWidth/2+rand.Intn(e.set.MountainRangeWidth),
						centre.Y-e.set.MountainRangeWidth/2+rand.Intn(e.set.MountainRangeWidth),
					)
					cnv.Ellipse(
						cp,
						e.set.Mountain.Roll(),
						e.set.Mountain.Roll(),
						rand.Intn(90),
						(rand.Float64()/4+rangeHeight)*scale,
						paint.Convex,
					)
					placed = append(placed, cp)
				}
				p += 1 + e.set.MountainStep.Roll()
			}
		}

		errchan <- op.pnt.Save(cnv)
	}()

	return placed, path, fanIn(errchan, wg)
}
