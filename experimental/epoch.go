package main

import (
	"github.com/voidshard/genesis"
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

	err = gen.NextEpoch(p.ID)
	if err != nil {
		panic(err)
	}
}
