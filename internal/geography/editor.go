package geography

import (
	"fmt"
	"image"

	"github.com/voidshard/genesis/internal/config"
	"github.com/voidshard/genesis/internal/database"
	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/internal/voronoi"
	"github.com/voidshard/genesis/pkg/types"
)

const (
	// tags for features & canvas names
	tagMountains  = "mountains"
	tagVolcanoes  = "volcanoes"
	tagRavines    = "ravines"
	tagRivers     = "rivers"
	tagSea        = "sea"
	tagLand       = "land"
	tagRain       = "rain"
	tagPerlin     = "noise-perlin"  // nice smooth noise
	tagVoro       = "noise-voronoi" // rough fractal style noise
	tagSeaCurrent = "sea-current"
)

var (
	// ErrNoPath returns if we cannot find a path between two points
	ErrNoPath = fmt.Errorf("failed to find valid path")

	// weights undrstood by out voronoi diagram implementation
	voroWeights = []string{
		tagMountains,
		tagRavines,
		tagRivers,
		tagVolcanoes,
		tagLand,
		tagSea,
		tagSeaCurrent,
	}

	// stuff we cart over
	copyCanvasBetweenEpoch = []string{
		tagMountains,
		tagRavines,
		tagRivers,
		tagVolcanoes,
		tagSea,
		tagRain,
	}
)

type Editor struct {
	cfg *config.Config
	db  database.Database

	proj  *types.Project
	graph voronoi.Graph
	hmap  map[image.Rectangle]image.Image

	set *Settings
}

func New(cfg *config.Config, db database.Database, set *Settings) *Editor {
	return &Editor{
		cfg:  cfg,
		db:   db,
		set:  set,
		hmap: map[image.Rectangle]image.Image{},
	}
}

//
func (e *Editor) NextEpoch(id string) error {
	p, err := e.project(id)
	if err != nil {
		return err
	}

	// copy canvases across to new epoch
	pnt := paint.New(e.cfg.Gen.Root, p.WorldWidth, p.WorldHeight)
	for _, n := range copyCanvasBetweenEpoch {
		co, err := pnt.Canvas(p.Canvas(n))
		if err != nil {
			return err
		}

		cn, err := pnt.NewCanvasFromImage(p.CanvasFromEpoch(n, p.Epoch+1), co.Image())
		if err != nil {
			return err
		}

		err = pnt.Save(cn)
		if err != nil {
			return err
		}
	}

	// increment project epoch
	p.Epoch += 1
	tx, err := e.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.SetProjects([]*types.Project{p})
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// project gets a single project by ID, proj is cached
func (e *Editor) project(id string) (*types.Project, error) {
	if e.proj != nil && e.proj.ID == id {
		return e.proj, nil
	}
	found, err := e.db.Projects([]string{id})
	if err != nil {
		return nil, err
	}
	if len(found) != 1 {
		return nil, fmt.Errorf("project %s not found", id)
	}
	return found[0], nil
}

func (e *Editor) cachedHeightmap(proj string, area image.Rectangle) (image.Image, error) {
	hmap, ok := e.hmap[area]
	if ok {
		return hmap, nil
	}
	hmap, err := e.HeightMap(proj, area)
	e.hmap[area] = hmap
	return hmap, err
}

func (e *Editor) cachedGraph(voro voronoi.Voronoi, name string) (voronoi.Graph, error) {
	if e.graph != nil {
		if e.graph.Name() == name {
			return e.graph, nil
		}
	}

	graph, err := voro.Graph(name)
	if err != nil {
		return nil, err
	}

	e.graph = graph
	return graph, nil
}
