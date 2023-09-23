package genesis

import (
	"log"
	"os"

	"github.com/voidshard/genesis/internal/config"
)

type Options struct {
	ConfigFile     string
	Root           string
	DatabaseDriver string
	SearchDriver   string
}

// setOpts sets options (from opts) in our config if they're non zero values
func setOpts(opts *Options, cfg *config.Config) {
	if opts.Root != "" {
		_, err := os.Stat(opts.Root)
		if os.IsNotExist(err) {
			err = os.MkdirAll(opts.Root, 0755)
			if err != nil {
				log.Println("failed to make folder", opts.Root, err)
			}
		}
		cfg.Gen.Root = opts.Root
	}
	if opts.DatabaseDriver != "" {
		cfg.Database.Driver = opts.DatabaseDriver
	}
	if opts.SearchDriver != "" {
		cfg.Search.Driver = opts.SearchDriver
	}
	if cfg.Database.Driver == config.DatabaseDriverSQLite {
		if cfg.Database.Location == "" {
			cfg.Database.Location = cfg.Gen.Root
		}
	}
	if cfg.Search.Driver == config.SearchDriverBleve {
		if cfg.Search.Location == "" {
			cfg.Search.Location = cfg.Gen.Root
		}
	}
	log.Println("using root folder", cfg.Gen.Root)
}
