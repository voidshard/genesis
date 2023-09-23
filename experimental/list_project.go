package main

import (
	"fmt"
	"github.com/voidshard/genesis/pkg/genesis"
)

const desc = ``

func main() {
	cli := &genesis.Options{Root: "/tmp/genesis"}

	gen, err := genesis.New(cli)
	if err != nil {
		panic(err)
	}

	found, tkn, err := gen.ListProjects("")
	if err != nil {
		panic(err)
	}
	pages := 0
	for {
		fmt.Println("page", pages)
		for i, proj := range found {
			fmt.Println(i, tkn, proj.ID, proj.Name)
		}
		if tkn == "" {
			break
		}
		pages += 1
		found, tkn, err = gen.ListProjects(tkn)
	}

}
