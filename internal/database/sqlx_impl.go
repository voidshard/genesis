package database

import (
	"github.com/voidshard/genesis/internal/dbutils"
	"github.com/voidshard/genesis/pkg/types"

	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	TableMeta       = "meta"
	TableProjects   = "projects"
	TableLandmasses = "landmasses"
	chunksize       = 6000
)

// sqlDB represents a generic DB wrapper -- this allows SQLite & Postgres to run
// the same code & execute queries in the same manner.
// We do have to be careful as the two underlying DBs are not exactly the same,
// but for our simple query requirements it's fine.
type sqlDB struct {
	conn *sqlx.DB
}

// Meta returns saved metadata, outside of a transaction
func (s *sqlDB) Meta(id string) (string, int, error) {
	return meta(s.conn, id)
}

// Projects fetches projects outside of a transaction
func (s *sqlDB) Projects(ids []string) ([]*types.Project, error) {
	return projects(s.conn, ids)
}

// Landmasses fetches landmass objects from the DB
func (s *sqlDB) Landmasses(ids []string) ([]*types.Landmass, error) {
	return landmasses(s.conn, ids)
}

// ListProjects iterates over the project table with some iter token
func (s *sqlDB) ListProjects(token string) ([]*types.Project, string, error) {
	return listProjects(s.conn, token)
}

// ListLandmasses iterates over landmasses belonging to the given project with some token
func (s *sqlDB) ListLandmasses(projectID string, token string) ([]*types.Landmass, string, error) {
	return listLandmasses(s.conn, projectID, token)
}

// Close connection to DB
func (s *sqlDB) Close() error {
	return s.conn.Close()
}

// Begin a new transaction.
// You *must* either call Commit() or Rollback() after calling this
func (s *sqlDB) Begin() (Transaction, error) {
	tx, err := s.conn.Beginx()
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx: tx}, nil
}

// sqlTx represents a transaction with read/write powers
type sqlTx struct {
	tx *sqlx.Tx
}

// Commit finish & commit changes to the database
func (t *sqlTx) Commit() error {
	return t.tx.Commit()
}

// Rollback abort transaction without changing anything
func (t *sqlTx) Rollback() error {
	return t.tx.Rollback()
}

// Meta returns saved metadata
func (t *sqlTx) Meta(id string) (string, int, error) {
	return meta(t.tx, id)
}

// SetMeta sets some metadata within a transaction
func (t *sqlTx) SetMeta(id, strv string, intv int) error {
	return setMeta(t.tx, id, strv, intv)
}

// Projects reads projects inside transaction
func (t *sqlTx) Projects(ids []string) ([]*types.Project, error) {
	return projects(t.tx, ids)
}

// SetProjects writes projects (insert-or-update) inside transaction
func (t *sqlTx) SetProjects(in []*types.Project) error {
	return setProjects(t.tx, in)
}

// SetLandmasses writes landmasses (insert or update) inside transaction
func (t *sqlTx) SetLandmasses(in []*types.Landmass) error {
	return setLandmasses(t.tx, in)
}

// Landmasses reads landmasses inside transaction
func (t *sqlTx) Landmasses(ids []string) ([]*types.Landmass, error) {
	return landmasses(t.tx, ids)
}

// DeleteLandmassesByProject removes all landmasses of the given project & epoch
func (t *sqlTx) DeleteLandmassesByProjectEpoch(projectID string, e int) error {
	return deleteLandmassesByProjectEpoch(t.tx, projectID, e)
}

// sqlOperator is something that can perform an sql operation read/write
// We do this so we can have some lower level funcs that perform the query logic regardless
// of whether we are in a transaction or not.
type sqlOperator interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	Select(dest interface{}, query string, args ...interface{}) error
}

// listProjects iterates over the projects table based on a token
func listProjects(op sqlOperator, tkn string) ([]*types.Project, string, error) {
	itr, err := dbutils.ParseIterToken(tkn)
	if err != nil {
		return nil, "", err
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s ORDER BY id LIMIT %d OFFSET %d;",
		TableProjects,
		itr.Limit,
		itr.Offset,
	)

	result := []*types.Project{}
	err = op.Select(&result, query)

	if err != nil {
		return nil, tkn, err
	} else if len(result) < itr.Limit {
		return result, "", nil
	} else {
		itr.Offset += itr.Limit
		return result, itr.String(), nil
	}
}

// listLandmasses iterates over landmasses belonging to a given project
func listLandmasses(op sqlOperator, projectID, tkn string) ([]*types.Landmass, string, error) {
	itr, err := dbutils.ParseIterToken(tkn)
	if err != nil {
		return nil, "", err
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s ORDER BY id LIMIT %d OFFSET %d;",
		TableLandmasses,
		itr.Limit,
		itr.Offset,
	)

	result := []*types.Landmass{}
	err = op.Select(&result, query)

	if err != nil {
		return nil, tkn, err
	} else if len(result) < itr.Limit {
		return result, "", nil
	} else {
		itr.Offset += itr.Limit
		return result, itr.String(), nil
	}
}

// mstruct is a row of metadata
type mstruct struct {
	ID  string `db:"id"`
	Str string `db:"str"`
	Int int    `db:"int"`
}

// meta returns some metadata, if set
func meta(op sqlOperator, id string) (string, int, error) {
	if !dbutils.IsValidName(id) {
		return "", 0, fmt.Errorf("metadata key %s is invalid", id)
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE id=$1 LIMIT 1;",
		TableMeta,
	)

	result := []*mstruct{}
	err := op.Select(&result, query, id)
	if err != nil || len(result) == 0 {
		return "", 0, err
	}

	return result[0].Str, result[0].Int, nil
}

// setMeta sets some data in our meta table
func setMeta(op sqlOperator, id, strv string, intv int) error {
	if !dbutils.IsValidName(id) {
		return fmt.Errorf("metadata key %s is invalid", id)
	}

	// update schema version to current
	qstr := fmt.Sprintf(`INSERT INTO %s (id, str, int)
		VALUES (:id, :str, :int) 
		ON CONFLICT (id) DO UPDATE SET
		    int=EXCLUDED.int,
		    str=EXCLUDED.str
		;`,
		TableMeta,
	)
	_, err := op.NamedExec(qstr, map[string]interface{}{
		"id":  id,
		"str": strv,
		"int": intv,
	})
	return err
}

// projects base level func to build & query for projects
func projects(op sqlOperator, ids []string) ([]*types.Project, error) {
	wstr, args := queryByIds(ids)
	if args == nil {
		return nil, nil
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s %s LIMIT %d;",
		TableProjects,
		wstr,
		len(ids),
	)

	result := []*types.Project{}
	return result, op.Select(&result, query, args...)
}

// setProjects base level func to write projects
func setProjects(op sqlOperator, in []*types.Project) error {
	for _, p := range in {
		if !dbutils.IsValidID(p.ID) {
			return fmt.Errorf("project id %s is invalid", p.ID)
		}
	}

	qstr := fmt.Sprintf(
		`INSERT INTO %s (id, name, epoch, seed, world_width, world_height)
		VALUES (:id, :name, :epoch, :seed, :world_width, :world_height) 
		ON CONFLICT (id) DO UPDATE SET
		    epoch=EXCLUDED.epoch,
		    seed=EXCLUDED.seed,
		    world_width=EXCLUDED.world_width,
		    world_height=EXCLUDED.world_height
		;`,
		TableProjects,
	)
	_, err := op.NamedExec(qstr, in)
	return err
}

// landmasses base level func to query landmasses
func landmasses(op sqlOperator, ids []string) ([]*types.Landmass, error) {
	wstr, args := queryByIds(ids)
	if args == nil {
		return nil, nil
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s %s LIMIT %d;",
		TableLandmasses,
		wstr,
		len(ids),
	)

	result := []*types.Landmass{}
	return result, op.Select(&result, query, args...)
}

// setLandmasses updates landmass objects in place
func setLandmasses(op sqlOperator, in []*types.Landmass) error {
	for _, l := range in {
		if !dbutils.IsValidID(l.ProjectID) {
			return fmt.Errorf("landmass project id %s is invalid", l.ProjectID)
		}
		if !dbutils.IsValidID(l.ID) {
			return fmt.Errorf("landmass id %s is invalid", l.ID)
		}
	}

	qstr := fmt.Sprintf(
		`INSERT INTO %s (project_id, id, size, color_r, color_g, color_b, first_x, first_y)
		VALUES (:project_id, :id, :size, :color_r, :color_g, :color_b, :first_x, :first_y) 
		ON CONFLICT (id) DO UPDATE SET
		    size=EXCLUDED.size,
		    color_r=EXCLUDED.color_r,
		    color_g=EXCLUDED.color_g,
		    color_b=EXCLUDED.color_b,
		    first_x=EXCLUDED.first_x,
		    first_y=EXCLUDED.first_y
		;`,
		TableLandmasses,
	)
	_, err := op.NamedExec(qstr, in)
	return err
}

func deleteLandmassesByProjectEpoch(op sqlOperator, projectID string, e int) error {
	if !dbutils.IsValidID(projectID) {
		return fmt.Errorf("project id %s is invalid", projectID)
	}
	_, err := op.NamedExec(
		fmt.Sprintf(`DELETE FROM %s WHERE project_id=:id AND epoch=:epoch;`, TableLandmasses),
		map[string]interface{}{"id": projectID, "epoch": e},
	)
	return err
}

func queryByIds(ids []string) (string, []interface{}) {
	if ids == nil || len(ids) == 0 {
		return "", nil
	}
	args := []interface{}{}
	or := []string{}
	for i, id := range ids {
		if !dbutils.IsValidID(id) {
			continue
		}
		name := fmt.Sprintf("$%d", i)
		args = append(args, id)
		or = append(or, fmt.Sprintf("id=:%s", name))
	}
	if len(or) == 0 {
		return "", nil
	}
	return fmt.Sprintf(" WHERE %s", strings.Join(or, " OR ")), args
}
