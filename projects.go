package genesis

import (
	"fmt"
	"math/rand"

	"github.com/voidshard/genesis/internal/dbutils"
	"github.com/voidshard/genesis/pkg/types"
)

const (
	minWorldSize = 500
)

var (
	// ErrNotFound our generic 404
	ErrNotFound = fmt.Errorf("not found")
)

// Project returns the given project by name or ID.
// We will assume ID first, otherwise Name.
func (e *Editor) Project(key string) (*types.Project, error) {
	id := key
	if !dbutils.IsValidID(key) {
		// possible because IDs are deterministic
		id = dbutils.NewID(key)
		fmt.Println("SET ID TO", id)
	}
	ps, err := e.Projects([]string{id})
	if err != nil {
		return nil, err
	}
	if len(ps) == 1 {
		return ps[0], nil
	}
	return nil, fmt.Errorf("%w project '%s'", ErrNotFound, key)
}

func (e *Editor) Projects(ids []string) ([]*types.Project, error) {
	return e.db.Projects(ids)
}

func (e *Editor) ListProjects(tkn string) ([]*types.Project, string, error) {
	return e.db.ListProjects(tkn)
}

func (e *Editor) CreateProject(in *types.Project) error {
	in.ID = dbutils.NewID(in.Name)

	// overwrite anything invalid
	if in.Seed <= 0 {
		in.Seed = int(rand.Int63())
	}
	if in.WorldWidth < minWorldSize {
		in.WorldWidth = minWorldSize
	}
	if in.WorldHeight < minWorldSize {
		in.WorldHeight = minWorldSize
	}

	txn, err := e.db.Begin()
	if err != nil {
		return err
	}
	err = txn.SetProjects([]*types.Project{in})
	if err != nil {
		txn.Rollback()
		return err
	}

	return txn.Commit()
}
