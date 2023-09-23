package voronoi

import (
	"image"

	"github.com/voidshard/voronoi"
)

//
type Cell struct {
	// Site is the centre of the cell
	Site image.Point

	// parent site in voronoi diagram implementation
	parent voronoi.Site
}

// ID returns unique number ID for cell
func (c *Cell) ID() int { return c.parent.ID() }

// Edges returns polygon that makes up the cell borders
func (c *Cell) Edges() [][2]image.Point { return c.parent.Edges() }
