package genesis

import (
	"image"

	"github.com/voidshard/genesis/pkg/types"
)

type GenesisEditor interface {
	projectEditor
	geographyEditor
	raceEditor
	civilizationEditor
}

type projectEditor interface {
	// CreateProject makes a new project
	CreateProject(*types.Project) error

	// ListProjects iterates over all projects
	ListProjects(tkn string) ([]*types.Project, string, error)

	// Projects returns projects by their ID(s)
	Projects([]string) ([]*types.Project, error)

	// Project returns a project by ID (sugar for 'Projects')
	Project(key string) (*types.Project, error)
}

type geographyEditorInit interface {
	// Tectonics divides the map into regions - used by following
	// functions that pick out paths between points.
	CreateTectonics(proj string, noise float64, points int) error
}

type geographyEditorTerrain interface {
	// Functions here imply
	// - CreateTectonics

	// A mountain range follows some path, placing high ridges and mountains
	// randomly along the path
	AddMountainRange(proj, tag string, s *types.PathSpec, scale float64) ([]image.Point, []image.Point, error)

	// Similar to mountain range we place volcanoes around a rough path
	// (eg. a fault line) but much less frequently than mountains.
	AddVolanoes(proj string, count int, s *types.PathSpec) ([]image.Point, []image.Point, error)

	// A ravine follows a path, adding steep sheer cliff walls
	AddRavine(proj, tag string, s *types.PathSpec, forkChance float64) ([]image.Point, error)

	// SmoothTerrain applies a smoothing brush to mountains / volcanoes
	SmoothTerrain(proj string, radius uint32) error

	// FlattenOutside terrain (eg.outside the rect) at the very edge(s) of the map down to 0
	// Ie. if you wished to force the edges to be sea .. this would be how
	FlattenOutside(proj string, r image.Rectangle) error
}

type geographyEditorDerived interface {
	// Functions here imply functions in 'geographyEditorTerrain' interface are called.
	// Since these functions use data written by them. This also implies if you go back and
	// recall them the outputs here could be out of date.

	// SeaMap figures out where there should be sea, sea temperatures (including currents).
	// We also this this time to figure out where land is (ie. not sea ..) and the size / location
	// of each landmass.
	SeaMap(proj string, sealevel uint8, equatorWidth, arcticWidth, seaCurrents int) (image.Image, []*types.Landmass, error)

	// HeightMap generates an amalgamated height map
	HeightMap(proj string, area image.Rectangle) (image.Image, error)

	// Rain determines rainfall & rainshadows.
	// Implies
	// - SeaMap
	Rain(proj string, stormMult float64, prevailingWinds []types.Heading) (image.Image, error)

	// Rivers determines where rivers should go based on rainfall.
	// Ie. Water flows downward & collects before returning to the sea.
	// Implies
	// - Rain
	Rivers(proj string, threshold int) (image.Image, error)
}

type geographyEditor interface {
	// functions to call when setting up an initial worldspace (once)
	geographyEditorInit

	// functions that alter terrain of the current epoch (per epoch, before 'Derived'
	geographyEditorTerrain

	// functions that use information from previous steps (init / terrain) and should
	// be (re)called after changes to them.
	geographyEditorDerived

	// NextEpoch moves us to a new epoch. Ie; the caller intends to apply
	// major land / sea / rain variation (effectively making a 'next' world
	// based on the current one) but wants to keep information from
	// this epoch.
	//
	// - updates current project epoch+1
	// - copies canvases (maps) for mountains, volcanoes, ravines into new epoch
	//   (that is, the current terrain)
	// - each epoch re-uses noise maps / tectonics
	// - since we're assuming the caller will add new mountains / volcanoes / ravines
	//   derived stuff like sea, rain, heightmap(s) will need re-calculation (that is,
	//   we don't copy derived information we expect will be outdated immediately)
	NextEpoch(proj string) error
}

type raceEditor interface {
}

type civilizationEditor interface {
}
