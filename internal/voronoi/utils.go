package voronoi

import (
	"image"
	"math"

	"github.com/voidshard/voronoi"
	"github.com/voidshard/voronoi/line"
)

const (
	siteMinDist = 25
)

// PointsBetween returns all points between a,b
func PointsBetween(a, b image.Point) []image.Point {
	return line.PointsBetween(a, b)
}

// pythagoras returns the dist between two points
func pythagoras(a, b image.Point) float64 {
	return math.Sqrt(math.Pow(float64(a.X+b.X), 2) + math.Pow(float64(a.Y+b.Y), 2))
}

// toDecimal turns floats without a decimal component into a number beginning
// with 0.
// Ie. 154 -> 0.154
//       3 -> 0.3
func toDecimal(f float64) float64 {
	count := 0
	tmp := int(f)
	for {
		if tmp <= 0 {
			break
		}
		count++
		tmp /= 10
	}
	return f / math.Pow10(count)
}

// rebuildVoronoi returns the diagram given it's sites.
//
// This is .. roughly equal to the one output by `randomVoronoi` assuming it
// has the same sites. That is, we use this one for cells, but the actual
// locations of the points can be off due to rounding errors in the
// implementation. So this one is considered useful for sites & rough calcs
// but the vertices / edges from the first call of randomVoronoi should be
// saved to ensure accuracy.
func rebuildVoronoi(width, height int, pts []image.Point) (*voronoi.Voronoi, error) {
	b := voronoi.NewBuilder(image.Rect(0, 0, width, height))
	for _, p := range pts {
		b.AddSite(p.X, p.Y)
	}
	return b.Voronoi()
}

// randomVoronoi returns a voronoi diagram with approximately `numPoints` Sites.
// We return the sites, vertices (corners of sites) and edges (edges between
// vertices).
func randomVoronoi(seed int64, width, height, points int) (*voronoi.Voronoi, []image.Point, []image.Point, [][2]image.Point, error) {
	b := voronoi.NewBuilder(image.Rect(0, 0, width, height))
	if seed > 0 {
		b.SetSeed(seed)
	}
	b.SetSiteFilters(b.MinDistance(siteMinDist))

	sites := []image.Point{}
	for i := 0; i < points*2; i++ {
		x, y, _, ok := b.AddRandomSite()
		if ok {
			sites = append(sites, image.Pt(x, y))
		}
		if b.SiteCount() >= points {
			break
		}
	}

	diagram, err := b.Voronoi()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	edgeId := func(a, b image.Point) float64 {
		// gets a unique ID for each edge regardless of point order
		aid := float64(a.Y*width + a.X)
		bid := float64(b.Y*width + b.X)
		if bid < aid {
			aid, bid = bid, aid
		}
		return aid + toDecimal(bid)
	}

	edgesSeen := map[float64]bool{}
	vertsSeen := map[image.Point]bool{}

	vertices := []image.Point{}
	edges := [][2]image.Point{}
	for _, s := range diagram.Sites() {
		for _, e := range s.Edges() {
			// save unique verts
			_, seenZro := vertsSeen[e[0]]
			_, seenOne := vertsSeen[e[1]]
			vertsSeen[e[0]] = true
			vertsSeen[e[1]] = true
			if !seenZro {
				vertices = append(vertices, e[0])
			}
			if !seenOne {
				vertices = append(vertices, e[1])
			}

			// save unique edges
			eid := edgeId(e[0], e[1])
			_, edgeSeen := edgesSeen[eid]
			edgesSeen[eid] = true
			if !edgeSeen {
				edges = append(edges, e)
			}
		}
	}

	return diagram, sites, vertices, edges, nil
}
