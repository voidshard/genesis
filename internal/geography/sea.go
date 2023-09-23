package geography

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/voidshard/genesis/internal/dbutils"
	"github.com/voidshard/genesis/internal/dijkstra"
	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/internal/voronoi"
	"github.com/voidshard/genesis/pkg/types"
)

// SeaMap figures out where the sea should go & cold/hot ocean water currents
func (e *Editor) SeaMap(proj string, sealevel uint8, equatorWidth, arcticWidth, currents int) (image.Image, []*types.Landmass, error) {
	p, err := e.project(proj)
	if err != nil {
		return nil, nil, err
	}

	voro := voronoi.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)
	graph, err := e.cachedGraph(voro, p.VoronoiDiagram())
	if err != nil {
		return nil, nil, err
	}

	hmap, err := e.HeightMap(proj, image.Rect(0, 0, p.WorldWidth, p.WorldHeight))
	if err != nil {
		return nil, nil, err
	}

	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)

	// determine what pixels are in the sea and which are not
	// note that this is simply a rough starting point ..
	sea, err := e.determineSea(p, pnt, hmap, voro, graph, sealevel)
	if err != nil {
		return nil, nil, err
	}

	// determine where ocean currents might run
	currentPaths, err := e.determineWaterCurrent(p, sea, graph.Points(), equatorWidth, arcticWidth, currents)
	if err != nil {
		return nil, nil, err
	}

	// now we can paint the actual sea, with equator, poles, currets etc
	sea, err = e.paintSea(p, pnt, sea, equatorWidth, arcticWidth, currentPaths)
	if err != nil {
		return nil, nil, err
	}

	// now that we know where the sea is, we can figure out the land
	landmasses, err := e.determineLand(p, pnt, sea)
	if err != nil {
		return nil, nil, err
	}

	// save landmasses (and flush old ones)
	tx, err := e.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	err = tx.DeleteLandmassesByProjectEpoch(p.ID, p.Epoch) // they've all changed (probably)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	err = tx.SetLandmasses(landmasses) // insert new landmasses
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	tx.Commit()

	return sea.Image(), landmasses, nil
}

// determineLand works out, once we've discovered where the sea goes, all of the unique landmasses.
//
// We track how large each landmass is, the first pixel we encountered and assign it a color.
// The number we use for the color is currently a uint16 (max 65k or so) so the color is not unique
// when the number of unique land forms exceeds that.
func (e *Editor) determineLand(proj *types.Project, pnt paint.Painter, sea paint.Canvas) ([]*types.Landmass, error) {
	seen := map[image.Point]bool{}
	bnds := sea.Bounds()

	sea.SetMask(nil)
	found := []*types.Landmass{}

	for dy := bnds.Min.Y; dy < bnds.Max.Y; dy++ {
		for dx := bnds.Min.X; dx < bnds.Max.X; dx++ {
			if sea.B(dx, dy) > 0 {
				continue
			}

			p := image.Pt(dx, dy)
			_, done := seen[p]
			if done {
				continue
			}

			// we've found a new landmass
			num := len(found) + 1
			if num > math.MaxUint16 {
				num -= math.MaxUint16 // wrap around
			}

			red, green := splitUint16(uint16(num)) // pick a color using red/green
			landColor := color.RGBA{red, green, 0, 255}

			size := 1
			stack := []image.Point{p} // neighbouring pixels to check
			for {
				if len(stack) < 1 {
					found = append(found, &types.Landmass{
						ProjectID: proj.ID,
						ID:        dbutils.RandomID(),
						Epoch:     proj.Epoch,
						Size:      size,
						ColorR:    int(red),
						ColorG:    int(green),
						FirstX:    p.X,
						FirstY:    p.Y,
					})
					break
				}

				next := stack[0]

				for px := next.X - 1; px <= next.X+1; px++ { // look at surrounding pixels
					if px < bnds.Min.X || px >= bnds.Max.X {
						continue
					}
					for py := next.Y - 1; py <= next.Y+1; py++ {
						if py < bnds.Min.Y || py >= bnds.Max.Y {
							continue
						}
						if px == next.X && py == next.Y {
							continue
						}
						if sea.B(px, py) > 0 {
							continue
						}

						candidate := image.Pt(px, py)
						_, ok := seen[candidate]
						if ok {
							continue
						}
						seen[candidate] = true

						size += 1
						sea.Set(px, py, landColor)
						stack = append(stack, candidate)
					}
				}

				stack[0] = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			}
		}
	}

	return found, pnt.Save(sea)
}

// determineSea; figure out where the sea should be.
//
// By definition the sea is everything below sea level .. almost. Technically land can be below
// sea level if it doesn't adjoin somewhere that is also sea. So we check all map edges and
// label anything below sea level as sea. Then all adjoining pixels that are *also* below
// sea level are sea.
//
// At the same time, we weight the 'sea' and 'land' verticies of our voronoi diagram
// for use later (so we don't have to iterate over the oceans again ..)
func (e *Editor) determineSea(p *types.Project, pnt paint.Painter, hmap image.Image, voro voronoi.Voronoi, graph voronoi.Graph, sealevel uint8) (paint.Canvas, error) {
	onGraph := map[image.Point]bool{}
	for _, p := range graph.Points() {
		onGraph[p] = true
	}

	seaColor := color.RGBA{0, 0, 255, 255}

	sea, err := pnt.Canvas(p.Canvas("sea-temporary")) // temporary sea (used for mask, not saved)
	if err != nil {
		return nil, err
	}

	u8 := func(in uint32) uint8 {
		return uint8(in >> 8)
	}

	// search around edges of the map for "sea"
	stack := []image.Point{}
	seen := map[image.Point]bool{}

	for dx := 0; dx < p.WorldWidth; dx++ {
		r0, _, _, _ := hmap.At(dx, 0).RGBA()
		if u8(r0) <= sealevel {
			i := image.Pt(dx, 0)
			seen[i] = true
			sea.Set(i.X, i.Y, seaColor)
			stack = append(stack, i)
		}

		r1, _, _, _ := hmap.At(dx, p.WorldHeight-1).RGBA()
		if u8(r1) <= sealevel {
			i := image.Pt(dx, p.WorldHeight-1)
			seen[i] = true
			sea.Set(i.X, i.Y, seaColor)
			stack = append(stack, i)
		}
	}
	for dy := 0; dy < p.WorldHeight; dy++ {
		r0, _, _, _ := hmap.At(0, dy).RGBA()
		if u8(r0) <= sealevel {
			i := image.Pt(0, dy)
			seen[i] = true
			sea.Set(i.X, i.Y, seaColor)
			stack = append(stack, i)
		}

		r1, _, _, _ := hmap.At(p.WorldWidth-1, dy).RGBA()
		if u8(r1) <= sealevel {
			i := image.Pt(p.WorldWidth-1, dy)
			seen[i] = true
			sea.Set(i.X, i.Y, seaColor)
			stack = append(stack, i)
		}
	}

	// find all adjoining pixels
	weightLand := []image.Point{} // land verticies to weight
	weightSea := []image.Point{}  // sea verticies to weight
	for {
		if len(stack) <= 0 {
			break
		}

		next := stack[0]

		for px := next.X - 1; px <= next.X+1; px++ {
			if px < 0 || px >= p.WorldWidth {
				continue // out of bounds
			}

			for py := next.Y - 1; py <= next.Y+1; py++ {
				if py < 0 || py >= p.WorldHeight {
					continue // out of bounds
				}
				if px == next.X && py == next.Y {
					continue // looking at myself
				}

				candidate := image.Pt(px, py)
				_, ok := seen[candidate]
				if ok {
					// point has already been checked
					continue
				}
				seen[candidate] = true

				_, applyWeight := onGraph[candidate]

				r, _, _, _ := hmap.At(px, py).RGBA()
				if u8(r) > sealevel {
					if applyWeight {
						weightLand = append(weightLand, candidate)
					}
					continue
				}

				sea.Set(candidate.X, candidate.Y, seaColor)
				stack = append(stack, candidate)
				if applyWeight {
					weightSea = append(weightSea, candidate)
				}
			}
		}

		stack[0] = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
	}

	/*
		// we weight things inversely;
		// SEA verticies on the 'land' graph are weighted heavily
		// LAND verticies on the 'sea' graph are weighted heavily
		err = graph.IncrWeights(weightSea, map[string]int{tagLand: e.set.GraphLandWeight})
		if err != nil {
			return nil, err
		}
		err = graph.IncrWeights(weightLand, map[string]int{tagSea: e.set.GraphSeaWeight})
		if err != nil {
			return nil, err
		}
	*/

	return sea, nil //, voro.Save(graph)
}

func (e *Editor) paintSea(p *types.Project, pnt paint.Painter, sea paint.Canvas, eqW, arW int, currents [][]image.Point) (paint.Canvas, error) {
	waterVCold := color.RGBA{0, 0, e.set.OceanWaterVeryCold, 255}
	waterCold := color.RGBA{0, 0, e.set.OceanWaterCold, 255}
	waterWarm := color.RGBA{0, 0, e.set.OceanWaterWarm, 255}
	waterVWarm := color.RGBA{0, 0, e.set.OceanWaterVeryWarm, 255}

	cur, err := pnt.Canvas(p.Canvas(tagSea))
	if err != nil {
		return nil, err
	}

	// use temp. sea image as a mask so we can't draw over land at all
	cur.SetMask(sea)

	// start with the Northern & Southern hemisphere which gives the rough temp. without currents
	eqTop := p.WorldHeight/2 - eqW/2
	eqBot := p.WorldHeight/2 + eqW/2
	cur.RectangleVertical(image.Rect(0, arW, p.WorldWidth, eqTop), waterCold, waterWarm, waterWarm)
	cur.RectangleVertical(image.Rect(0, eqBot, p.WorldWidth, p.WorldHeight-arW), waterWarm, waterWarm, waterCold)

	// paint in sea currents
	for _, path := range currents {
		if rand.Float64() <= e.set.OceanColdCurrentProb { // cold
			cur.Line(
				path,
				e.set.OceanCurrentWidth,
				waterVWarm,
				waterWarm,
				waterCold,
				waterCold,
				waterCold,
				waterVCold,
			)
		} else { // hot
			cur.Line(
				path,
				e.set.OceanCurrentWidth,
				waterVWarm,
				waterWarm,
				waterWarm,
				waterWarm,
				waterCold,
				waterVCold,
			)
		}
	}

	// draw in equator, Northern & Southern poles whose temperatures don't vary too much
	cur.RectangleVertical(image.Rect(0, eqTop, p.WorldWidth, eqBot), waterWarm, waterVWarm, waterWarm)
	cur.RectangleVertical(image.Rect(0, 0, p.WorldWidth, arW), waterVCold, waterCold)
	cur.RectangleVertical(image.Rect(0, p.WorldHeight-arW, p.WorldWidth, p.WorldHeight), waterCold, waterVCold)

	return cur, pnt.Save(cur)
}

// determineWaterCurrent figures out the temperature of the ocean, mostly we're interested in where
// warm & cold ocean currents are - since they influence later calculations on rainfall and/or
// lack there of.
func (e *Editor) determineWaterCurrent(p *types.Project, sea paint.Canvas, allPoints []image.Point, eqW, arW, currents int) ([][]image.Point, error) {
	isSea := func(x, y int) bool {
		_, _, b, _ := sea.At(x, y).RGBA()
		return b > 0
	}

	// we have a voronoi diagram .. but we're going to build another graph because;
	// 1. we want all points to be in the sea
	// 2. we only consider edges valid if they don't hit land (ocean currents don't traverse land...)
	// 3. we want sea currents to start in the equator & flow towards one of the poles
	//    (or the inverse)
	// In order to cut down on the checks we have to do we'll bucket sea points into
	// grid regions. Each grid section's points can be heavily interconnected, but we'll
	// limit connections between point(s) in a grid and neighbouring grid(s).
	// This means we at worst only check a point vs. points relatively close to it
	// (ie. within the same grid or the next door grids) and usually (if our inter-grid connectivity
	// is kept low) cut down on even that.
	points := []image.Point{}
	grid := NewGridBucket(p.WorldWidth, p.WorldHeight, e.set.OceanCurrentGridSize, false)
	north := []int{}
	south := []int{}
	centr := []int{}

	eqTop := p.WorldHeight/2 - eqW/2
	eqBot := p.WorldHeight/2 + eqW/2
	for _, v := range allPoints {
		if !isSea(v.X, v.Y) {
			continue
		}

		i := len(points)
		points = append(points, v)

		grid.AddPoint(i, v)
		if v.Y <= arW {
			north = append(north, i)
		} else if v.Y >= p.WorldHeight-arW {
			south = append(south, i)
		} else if v.Y >= eqTop && v.Y <= eqBot {
			centr = append(centr, i)
		}
	}

	edgeChan := grid.BuildEdges(
		func(a, b int) float64 {
			for _, p := range voronoi.PointsBetween(points[a], points[b]) {
				if !isSea(p.X, p.Y) {
					return -1 // reject this edge
				}
			}
			return 0 // accept this edge
		},
		func(a, b int) float64 {
			for _, p := range voronoi.PointsBetween(points[a], points[b]) {
				if !isSea(p.X, p.Y) {
					return -1 // reject this edge
				}
			}
			return distBetween(points[a].X, points[a].Y, points[b].X, points[b].Y)
		},
		2, // how many connections we allow between two grids
	)
	edges := [][2]image.Point{}
	for e := range edgeChan {
		edges = append(edges, [2]image.Point{points[e[0]], points[e[1]]})
	}

	// now with a list of points & edges between points, we can finally build a graph
	seaGraph, err := dijkstra.New(0, []string{tagSeaCurrent}, points, edges)
	if err != nil {
		return nil, err
	}

	seaCurrents := [][]image.Point{} // equator -> pole
	currentWeight := map[string]int{tagSeaCurrent: 10}
	for _, equator := range centr {
		for _, arctic := range north {
			curpath, err := seaGraph.Shortest(tagSeaCurrent, points[equator], points[arctic])
			if err != nil {
				continue // graph probably disconnected
			}
			if len(curpath) < 2 {
				continue
			}
			seaCurrents = append(seaCurrents, curpath)
			seaGraph.IncrWeights(curpath, currentWeight)
		}
		for _, arctic := range south {
			curpath, err := seaGraph.Shortest(tagSeaCurrent, points[equator], points[arctic])
			if err != nil {
				continue // graph probably disconnected
			}
			if len(curpath) < 2 {
				continue
			}
			seaCurrents = append(seaCurrents, curpath)
			seaGraph.IncrWeights(curpath, currentWeight)
		}
	}

	if len(seaCurrents) > currents {
		rand.Shuffle(len(seaCurrents), func(i, j int) {
			seaCurrents[i], seaCurrents[j] = seaCurrents[j], seaCurrents[i]
		})
		return seaCurrents[:currents], nil
	}

	return seaCurrents, nil
}
