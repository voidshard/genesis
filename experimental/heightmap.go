package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math/rand"
	"os"
	"sort"

	"github.com/voidshard/genesis"
	"github.com/voidshard/genesis/internal/paint"
	"github.com/voidshard/genesis/internal/voronoi"
	"github.com/voidshard/genesis/pkg/types"
)

const (
	numMountains = 20
	numVolcs     = 4
	numRivers    = 30
	numRavines   = 1

	perinNoiseVal = 0.07
	voronoiPoints = 1000
	smoothPasses  = 4

	sealevel uint8 = 150
)

func main() {
	cli := &genesis.Options{Root: "/tmp/genesis"}
	gen, err := genesis.New(cli)
	if err != nil {
		panic(err)
	}

	p, err := gen.Project("my-project")
	if err != nil {
		panic(err)
	}
	fmt.Println(p.WorldWidth, p.WorldHeight)

	log.Println("Cleaning previous data")
	pnt := paint.New(cli.Root, p.WorldWidth, p.WorldHeight)
	for _, t := range []string{"mountains", "ravines", "rivers", "sea", "noise-perlin", "noise-voronoi", "rain"} {
		pnt.Delete(p.Canvas(t))
	}

	voro := voronoi.New(cli.Root, p.WorldWidth, p.WorldHeight)
	voro.Delete(p.VoronoiDiagram())
	log.Println("Regenerating")

	err = gen.CreateTectonics(p.ID, perinNoiseVal, voronoiPoints)
	if err != nil {
		panic(err)
	}
	log.Println("Tectonics created")

	for i := 0; i < numMountains; i++ {
		_, _, err = gen.AddMountainRange(p.ID, "mountains-of-madness", nil, 1)
		if err != nil {
			panic(err)
		}
		log.Println("Added mountain range")
	}
	for i := 0; i < numVolcs; i++ {
		_, _, err = gen.AddVolanoes(p.ID, 5, nil)
		if err != nil {
			panic(err)
		}
		log.Println("Added volcanic region")
	}

	for i := 0; i < numRavines; i++ {
		_, err = gen.AddRavine(p.ID, "ravine", nil, .2)
		if err != nil {
			panic(err)
		}
		log.Println("Added ravine")
	}

	log.Println("\tSmoothing ...")
	for i := 0; i < smoothPasses; i++ {
		err = gen.SmoothTerrain(p.ID, uint32(rand.Intn(2)+rand.Intn(2)+2))
	}

	log.Println("Flattening edges ...")
	err = gen.FlattenOutside(p.ID, image.Rect(5, 5, p.WorldWidth-5, p.WorldHeight-5))
	if err != nil {
		panic(err)
	}

	hm, err := gen.HeightMap(p.ID, image.Rect(0, 0, p.WorldWidth, p.WorldHeight))
	if err != nil {
		panic(err)
	}
	log.Println("Completed heightmap")

	err = SavePNG("/tmp/genesis/heightmap.png", hm)
	if err != nil {
		panic(err)
	}

	log.Println("Discovering sea")
	_, land, err := gen.SeaMap(p.ID, sealevel, 100, 100, 6)
	if err != nil {
		panic(err)
	}
	sort.Slice(land, func(i, j int) bool {
		return land[i].Size > land[j].Size
	})
	fmt.Println("\tfound", len(land), "landmasses")
	for _, l := range land {
		fmt.Println("\t\t", l.ID, l.Size, l.FirstX, l.FirstY)
	}

	log.Println("Raining ..")
	_, err = gen.Rain(p.ID, 3, nil)
	if err != nil {
		panic(err)
	}
	_, err = gen.Rain(p.ID, 1, types.ClockwiseHeadings(gen.Geo.RainfallPrevailingWinds...))
	if err != nil {
		panic(err)
	}
	_, err = gen.Rain(p.ID, 1, types.CounterClockwiseHeadings(gen.Geo.RainfallPrevailingWinds...))
	if err != nil {
		panic(err)
	}
}

func SavePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}
