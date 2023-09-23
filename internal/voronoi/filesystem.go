package voronoi

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type fsVoronoi struct {
	root   string
	width  int
	height int
}

func New(root string, width, height int) Voronoi {
	os.MkdirAll(root, 0750)
	return &fsVoronoi{root: root, width: width, height: height}
}

func (f *fsVoronoi) NewGraph(name string, weightNames []string, defaultWeight, points int, seed int64) (Graph, error) {
	return newGraph(name, weightNames, f.width, f.height, defaultWeight, points, seed)
}

func (f *fsVoronoi) Graph(name string) (Graph, error) {
	data, err := ioutil.ReadFile(f.pathFor(name))
	if err != nil {
		return nil, err
	}
	g := &graph{}
	return g, g.Unmarshal(data)
}

func (f *fsVoronoi) Delete(name string) error {
	return os.Remove(f.pathFor(name))
}

func (f *fsVoronoi) Save(in Graph) error {
	data, err := in.Marshal()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(f.pathFor(in.Name()), data, 0660)
}

// pathFor retrns where on the disk we store this named graph
func (f *fsVoronoi) pathFor(name string) string {
	return filepath.Join(f.root, fmt.Sprintf("%s.json", name))
}
