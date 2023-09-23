package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/wlevene/ini"
)

const (
	// configEnvVar is what we check to find a config file
	configEnvVar = "GENESIS_CONFIG"

	// default config file name
	configFile = "config.ini"

	// folders that we'll check for our data (eg $HOME/.config/vendor/app/)
	vendor = "voidshard"
	app    = "genesis"
)

const (
	SearchDriverBleve = "bleve"

	DatabaseDriverSQLite = "sqlite3"
)

var (
	// the default config settings
	initial = []byte(`
[gen]
root=

[database]
driver=sqlite3
name=genesis.sqlite

[search]
driver=bleve
name=genesis.bleve

[geography]

[civilization]
`)
)

type Config struct {
	Gen struct {
		Root string `ini:"root"`
	} `ini:"gen"`

	Database Database `ini:"database"`

	Search Search `ini:"search"`

	Geography    Geography    `ini:"geography"`
	Civilization Civilization `ini:"civilization"`
}

type Database struct {
	Driver   string `ini:"driver"`
	Name     string `ini:"name"`
	Location string `ini:"location"` // host:port or folder
}

type Search struct {
	Driver   string `ini:"driver"`
	Name     string `ini:"name"`
	Location string `ini:"location"` // host:port or folder
}

type Geography struct {
}

type Civilization struct {
}

//
func New(explicit string) (*Config, error) {
	cfg := &Config{}

	// ensure the root is always set, default depends on OS
	defer func() {
		cfg.Gen.Root = rootFolder(cfg.Gen.Root)
	}()

	fpath := explicit
	if fpath == "" {
		// if not given check env var
		fpath = os.Getenv(configEnvVar)
	}
	if fpath == "" {
		log.Println("using default config")
		// if still not found use canned defaults
		return cfg, ini.Unmarshal(initial, cfg)
	}

	// otherwise, read file & return config
	log.Println("using config", fpath)
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return cfg, ini.Unmarshal(data, cfg)
}

// rootFolder - we'll try very hard to find ourselves a home
func rootFolder(desired string) string {
	return firstFolder(
		desired,
		filepath.Join(defaultHomeFolder(), fmt.Sprintf(".%s", app)),
		filepath.Join(os.TempDir(), app),
	)
}

//defaultHomeFolder returns the home dir if known
func defaultHomeFolder() string {
	f, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return f
}

// exists returns if the given path exists
func exists(in string) bool {
	_, err := os.Stat(in)
	return !os.IsNotExist(err)
}

// FirstFolder returns the first in this list of available folders.
// If they do not exist we attempt to create them.
// If not set (ie: "") or we fail to create we move on to the next.
// If all else fails we return the os TempDir.
func firstFolder(in ...string) string {
	for _, folder := range in {
		if folder == "" {
			continue
		}

		if exists(folder) {
			return folder
		}

		err := os.MkdirAll(folder, 0755)
		if err != nil {
			log.Println("failed to make folder", folder, err)
			continue
		}
		return folder
	}
	return os.TempDir()
}
