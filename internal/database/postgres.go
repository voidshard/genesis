package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/voidshard/genesis/internal/config"
)

type Postgres struct {
	*sqlDB
}

func NewPostgres(cfg *config.Database) (*Postgres, error) {
	db, err := sqlx.Connect("postgres", cfg.Location)
	me := &Postgres{&sqlDB{conn: db}}
	return me, err
}
