package search

import (
	"github.com/voidshard/genesis/internal/config"
)

type Search interface{}

func New(opts config.Search) (Search, error) {
	return nil, nil
}
