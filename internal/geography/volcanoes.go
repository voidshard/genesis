package geography

import (
	"fmt"
	"image"
	"math/rand"

	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/internal/voronoi"
	"github.com/voidshard/genesis/pkg/types"
)

// Volcanoes are similar to mountains in that they follow fault lines, but are placed
// less frequently & further out (they don't sit directly on the line).
func (e *Editor) AddVolanoes(proj string, count int, s *types.PathSpec) ([]image.Point, []image.Point, error) {
	op, err := e.newGraphOp(proj, s)
	if err != nil {
		return nil, nil, err
	}

	cnv, err := op.pnt.Canvas(op.p.Canvas(tagMountains))
	if err != nil {
		return nil, nil, err
	}

	path, err := op.graph.Shortest(tagVolcanoes, op.from, op.to)
	if err != nil {
		return nil, nil, fmt.Errorf("%w %v", ErrNoPath, err)
	}
	path = trimPath(path, op.maxDist)
	if len(path) < 2 {
		return nil, nil, fmt.Errorf("%w path too short", ErrNoPath)
	}

	// choose where we might put a volcano
	candidates := []image.Point{}
	for j := 1; j < len(path); j++ { // for each segment of the range
		alongLine := voronoi.PointsBetween(path[j-1], path[j])
		for p := rand.Intn(5); p < len(alongLine); {
			candidates = append(
				candidates,
				pointNear(alongLine[p], e.set.VolcanoRangeWidth/2, e.set.VolcanoRangeWidth),
			)
			p += 1 + e.set.VolcanoStep.Roll()
		}
	}

	if len(candidates) > count {
		rand.Shuffle(len(candidates), func(a, b int) {
			candidates[a], candidates[b] = candidates[b], candidates[a]
		})
		candidates = candidates[0:count]
	}
	for _, p := range candidates {
		cnv.Ellipse( // cone
			p,
			e.set.VolcanoCone.Roll(),
			e.set.VolcanoCone.Roll(),
			rand.Intn(90),
			0.75+rand.Float64()/4,
			paint.Convex,
		)
		cnv.Ellipse( // caldera
			p,
			e.set.VolcanoCaldera.Roll(),
			e.set.VolcanoCaldera.Roll(),
			rand.Intn(90),
			0.75+rand.Float64()/4,
			paint.Concave,
		)
	}

	return candidates, path, op.pnt.Save(cnv)
}
