package database

import (
	"github.com/voidshard/genesis/internal/config"
	"github.com/voidshard/genesis/pkg/types"
)

// Database is a top level interface for db interations
type Database interface {
	Read
	Iterate

	Close() error
	Begin() (Transaction, error)
}

// Transaction is required to write (update) entries.
// Note that Iterate is not part of the transaction interface.
type Transaction interface {
	Read
	Write

	Commit() error
	Rollback() error
}

// Iterate runs across a table, each call returns some iter token
// that will fetch the next page.
//
// This token, even if understandable, is internal only and should
// not be parsed -- it could change between releases without notice.
//
// Not permitted in a transaction
type Iterate interface {
	ListProjects(token string) ([]*types.Project, string, error)
	ListLandmasses(projectID string, token string) ([]*types.Landmass, string, error)
}

// Read allows one to look up items by their IDs
type Read interface {
	Projects([]string) ([]*types.Project, error)
	Meta(string) (string, int, error)
	Landmasses([]string) ([]*types.Landmass, error)
}

// Write updates the database, only usable in a Transaction
type Write interface {
	SetProjects([]*types.Project) error
	SetMeta(id, str_value string, int_value int) error
	SetLandmasses([]*types.Landmass) error
	DeleteLandmassesByProjectEpoch(id string, e int) error
}

// New returns a new database from a config
func New(opts config.Database) (Database, error) {
	if opts.Driver == config.DatabaseDriverSQLite {
		return NewSqlite3(&opts)
	}
	return nil, nil
}
