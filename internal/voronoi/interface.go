package voronoi

import (
	"image"
)

// Voronoi provides a database like interface for interacting with
// voronoi diagrams.
type Voronoi interface {
	// NewGraph makes a new graph and returns it.
	NewGraph(name string, wieghtNames []string, defaultWeight, points int, seed int64) (Graph, error)

	// Graph returns existing graph
	Graph(name string) (Graph, error)

	// Deletes graph. There is no error for attempting
	// to delete a graph that doesn't exist
	Delete(name string) error

	// Write out the graph
	Save(Graph) error
}

//
type Graph interface {
	// Name returns unique name of graph
	Name() string

	// Marshal graph into byte representation
	Marshal() ([]byte, error)

	// Unmarshal graph from bytes
	Unmarshal([]byte) error

	// RandomPoint returns a point at random from the graph
	RandomPoint() image.Point

	// RandomCell returns a voronoi diagram cell at random
	RandomCell() *Cell

	// ClosestPoint returns the closest point on the graph to
	// the given point.
	// Warning; calculating this could be expensive ..
	ClosestPoint(image.Point) image.Point

	// Points returns all points on the graph
	// These are vertexes that run alongside sites.
	//
	// Eg. given a site s1, the centre of a voronoi cell,
	//
	//   p1 ---------- p2 ---->
	//    \            |
	//     \     s1    |
	//      \          |    s2
	//       \         |
	//       p3------- p4 ----->
	//
	Points() []image.Point

	// Sites is the list of cell centres
	Sites() []image.Point

	// Shortest finds the path with the least weight in the
	// graph. Given points mapped to closest points in graph.
	Shortest(weight string, a, b image.Point) ([]image.Point, error)

	// IncrWeights for the given set of points.
	IncrWeights(pts []image.Point, delta map[string]int) error

	// IncrWeightsOutside is like IncrWeights but applies only to points
	// outside of the given bounds
	IncrWeightsOutside(area image.Rectangle, delta map[string]int) error

	// Tag tags all points in `p` with some name
	Tag(tag string, p []image.Point)

	// FromTag returns points tagged with `tag`
	FromTag(tag string) ([]image.Point, bool)

	// NeighbouringPoints returns all points that are joined to one of
	// the given points by an edge(s), but are *not* in `p`
	NeighbouringPoints(p []image.Point) ([]image.Point, error)

	// NeighbouringCells returns all nearby cells that share at least
	// one edge with the given cell(s) by are not in `c`
	NeighbouringCells(c []*Cell) ([]*Cell, error)
}
