package main

import (
	"github.com/alecthomas/kong"
	"github.com/voidshard/genesis/pkg/genesis"
	"github.com/voidshard/genesis/pkg/types"
)

const desc = ``

func main() {
	cli := &genesis.Options{}
	kong.Parse(cli, kong.Name("genesis"), kong.Description(desc))

	gen, err := genesis.New(cli)
	if err != nil {
		panic(err)
	}

	p := types.NewProject()
	p.Name = "my-project"

	err = gen.CreateProject(p)
	if err != nil {
		panic(err)
	}
	fmt.Println("project created", p.ID, p.Name)

	err = gen.SetDefaultProject(p.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("set project", p.ID)

	found, err := gen.Projects([]string{p.ID})
	if err != nil {
		panic(err)
	}
	fmt.Println("projects", found)
}
