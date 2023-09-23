package genesis

import (
	"log"

	"github.com/voidshard/genesis/internal/config"
	"github.com/voidshard/genesis/internal/database"
	"github.com/voidshard/genesis/internal/geography"
	"github.com/voidshard/genesis/internal/search"
)

// Check Editor implements GenesisEditor
var _ GenesisEditor = new(Editor)

// Editor is our top level struct
type Editor struct {
	cfg *config.Config
	db  database.Database
	sb  search.Search

	Geo     *geography.Settings
	geoEdit *geography.Editor
}

//
func New(opts *Options) (*Editor, error) {
	cfg, err := config.New(opts.ConfigFile)
	if err != nil {
		return nil, err
	}
	setOpts(opts, cfg)
	log.Println(cfg.Database, cfg.Search)

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}

	sb, err := search.New(cfg.Search)
	if err != nil {
		return nil, err
	}

	gs := geography.DefaultSettings()

	return &Editor{
		cfg:     cfg,
		db:      db,
		sb:      sb,
		Geo:     gs,
		geoEdit: geography.New(cfg, db, gs),
	}, nil
}
