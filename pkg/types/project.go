package types

import (
	"fmt"
)

type Project struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Epoch       int    `db:"epoch"`
	Seed        int    `db:"seed"`
	WorldWidth  int    `db:"world_width"`
	WorldHeight int    `db:"world_height"`
}

func (p *Project) Canvas(name string) string {
	// sugar for "canvas from the current epoch"
	return p.CanvasFromEpoch(name, p.Epoch)
}

func (p *Project) CanvasFromEpoch(name string, e int) string {
	return fmt.Sprintf("%s-%d-%s", p.ID, e, name)
}

func (p *Project) VoronoiDiagram() string {
	return fmt.Sprintf("%s-graph", p.ID)
}

func (p *Project) NoiseMap(i int) string {
	return fmt.Sprintf("%s-noise-%d", p.ID, i)
}

func NewProject() *Project {
	return &Project{}
}
