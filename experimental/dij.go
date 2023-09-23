package main

import (
	"fmt"
	"image"

	"github.com/voidshard/genesis/internal/graph"
)

func main() {
	pt0 := image.Pt(0, 0)
	pt1 := image.Pt(5, 5)
	pt2 := image.Pt(6, 6)
	pt3 := image.Pt(5, 10)

	g, err := graph.New(
		[]image.Point{pt0, pt1, pt2, pt3},
		[][2]image.Point{{pt0, pt1}, {pt1, pt2}, {pt3, pt1}, {pt2, pt3}},
		map[string][]int64{"a": nil, "b": nil},
	)
	fmt.Println(err, g)

	err = g.IncrWeights([]image.Point{pt0}, map[string]int64{"a": 11, "b": -1})

	fmt.Println(err, g)

	ns, err := g.Neighbours([]image.Point{image.Pt(5, 5), image.Pt(5, 10)})
	fmt.Println(err, ns)

	fmt.Println(g.RandomPoint())
	fmt.Println(g.RandomPoint())
	fmt.Println(g.RandomPoint())

	err = g.IncrWeightsOutside(image.Rect(1, 1, 9, 9), map[string]int64{"a": 100, "b": -50})
	fmt.Println(err, g)
	g.DW()
}
