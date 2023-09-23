package main

import (
	"fmt"
	"github.com/voidshard/genesis"
	"github.com/voidshard/genesis/pkg/types"
)

const desc = ``

func main() {
	cli := &genesis.Options{Root: "/tmp/genesis"}

	gen, err := genesis.New(cli)
	if err != nil {
		panic(err)
	}

	p := types.NewProject()
	p.Name = "my-project"
	p.WorldWidth = 1000
	p.WorldHeight = 1000

	err = gen.CreateProject(p)
	if err != nil {
		panic(err)
	}
	fmt.Println("project created", p.ID, p.Name)

	found, err := gen.Projects([]string{p.ID})
	if err != nil {
		panic(err)
	}
	fmt.Println("projects", found)
}
