package types

import (
	"image"
)

// PathSpec gives a rough outline for the path something should follow
type PathSpec struct {
	// From is where the path starts.
	// If not given a point will be randomly chosen.
	From *image.Point

	// To is where the path ends
	// If not given a point will be randomly chosen.
	To *image.Point

	// MaxDist is the approximate max distance the path should follow
	// starting from `From`
	MaxDist float64
}
