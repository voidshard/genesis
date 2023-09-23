package genesis

import (
	"image"

	"github.com/voidshard/genesis/pkg/types"
)

// Tectonics divides the map into regions - used by following
// functions that pick out paths between points.
func (e *Editor) CreateTectonics(proj string, noise float64, points int) error {
	return e.geoEdit.CreateTectonics(proj, noise, points)
}

//
func (e *Editor) Rain(proj string, stormMult float64, prevailingWinds []types.Heading) (image.Image, error) {
	return e.geoEdit.Rain(proj, stormMult, prevailingWinds)
}

//
func (e *Editor) Rivers(proj string, threshold int) (image.Image, error) {
	return e.geoEdit.Rivers(proj, threshold)
}

//
func (e *Editor) NextEpoch(proj string) error {
	return e.geoEdit.NextEpoch(proj)
}

// A mountain range follows some path, placing high ridges and mountains
// randomly along the path
// Implies
// - CreateTectonics
func (e *Editor) AddMountainRange(proj, tag string, s *types.PathSpec, scale float64) ([]image.Point, []image.Point, error) {
	return e.geoEdit.AddMountainRange(proj, tag, s, scale)
}

// Similar to mountain range we place volcanoes around a rough path
// (eg. a fault line) but much less frequently than mountains.
// Implies
// - CreateTectonics
func (e *Editor) AddVolanoes(proj string, count int, s *types.PathSpec) ([]image.Point, []image.Point, error) {
	return e.geoEdit.AddVolanoes(proj, count, s)
}

// A ravine follows a path, adding steep sheer cliff walls
// Implies
// - CreateTectonics
func (e *Editor) AddRavine(proj, tag string, s *types.PathSpec, forkChance float64) ([]image.Point, error) {
	return e.geoEdit.AddRavine(proj, tag, s, forkChance)
}

// SmoothTerrain applies a smoothing brush to mountains / volcanoes
func (e *Editor) SmoothTerrain(proj string, radius uint32) error {
	return e.geoEdit.SmoothTerrain(proj, radius)
}

// FlattenOutside terrain (eg.outside the rect) at the very edge(s) of the map down to 0
func (e *Editor) FlattenOutside(proj string, r image.Rectangle) error {
	return e.geoEdit.FlattenOutside(proj, r)
}

// SeaMap figures out where there should be sea.
// Implies
// - AddMountainRange
// - AddVolanoes
func (e *Editor) SeaMap(proj string, sealevel uint8, equatorWidth, articWidth, seaCurrents int) (image.Image, []*types.Landmass, error) {
	return e.geoEdit.SeaMap(proj, sealevel, equatorWidth, articWidth, seaCurrents)
}

// HeightMap generates an amalgamated height map using all of the previously
// called function(s) output.
// Implies
// - Anything that modifies terrain height .. obviously
func (e *Editor) HeightMap(proj string, area image.Rectangle) (image.Image, error) {
	return e.geoEdit.HeightMap(proj, area)
}
