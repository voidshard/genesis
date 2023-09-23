package database

import (
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/voidshard/genesis/internal/config"
)

const (
	// metaSchemaVersion is the current db schema version
	metaSchemaVersion = "db-schema-version"

	// currentSchemaVersion of the db schema. Should be updated
	// when we update the tables so we can handle migrations
	currentSchemaVersion = 1
)

var (
	createMeta = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	id VARCHAR(255) PRIMARY KEY,
	str TEXT NOT NULL DEFAULT "",
	int INTEGER NOT NULL DEFAULT 0
    );`, TableMeta)

	createProjects = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	id VARCHAR(255) PRIMARY KEY,
	name VARCHAR(255) UNIQUE NOT NULL DEFAULT "",
	epoch INTEGER NOT NULL DEFAULT 0,
	seed INTEGER NOT NULL DEFAULT 0,
	world_width INTEGER NOT NULL,
	world_height INTEGER NOT NULL
    );`, TableProjects)

	createLandmasses = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
	project_id VARCHAR(255) NOT NULL,
	id VARCHAR(255) PRIMARY KEY,
	epoch INTEGER NOT NULL DEFAULT 0,
	size INTEGER NOT NULL DEFAULT 0,
	color_r INTEGER NOT NULL DEFAULT 0,
	color_g INTEGER NOT NULL DEFAULT 0,
	color_b INTEGER NOT NULL DEFAULT 0,
	first_x INTEGER NOT NULL DEFAULT 0,
	first_y INTEGER NOT NULL DEFAULT 0
    );`, TableLandmasses)

	indexes = []string{}
)

// Sqlite represents a DB connection to sqlite
type Sqlite struct {
	*sqlDB
}

// NewSqlite3 opens a SQLite DB file.
// We also attempt to create / update tables to make it ready for use.
func NewSqlite3(cfg *config.Database) (*Sqlite, error) {
	db, err := sqlx.Connect("sqlite3", filepath.Join(cfg.Location, cfg.Name))
	if err != nil {
		return nil, err
	}
	me := &Sqlite{&sqlDB{conn: db}}
	return me, me.setupDatabase()
}

// setupDatabase tries to bring an SQLite db into line with our schema, upgrading as
// needed.
// We consider this acceptable in the case of local SQlite - but postgres is an entire
// 'nother kettle of fish.
func (s *Sqlite) setupDatabase() error {
	err := s.createTables()
	if err != nil {
		return err
	}

	txn, err := s.Begin()
	if err != nil {
		return err
	}

	err = txn.SetMeta(metaSchemaVersion, "", currentSchemaVersion)
	if err != nil {
		txn.Rollback()
		return err
	}

	return txn.Commit()
}

// createTables & wrap up an errors we get.
// We'll try to press on despite errors.
func (s *Sqlite) createTables() error {
	var final error
	todo := []string{createMeta, createProjects, createLandmasses}
	todo = append(todo, indexes...)
	for _, ddl := range todo {
		_, err := s.conn.Exec(ddl)
		if err != nil {
			if final == nil {
				final = err
			} else {
				final = fmt.Errorf("%w %v", final, err)
			}
		}
	}
	return final
}
