package database

import (
	"database/sql"
	"github.com/Whuichenggong/urlshortener/urlshortener/config"
	_ "github.com/lib/pq" //注入到sql
)

func NewDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN())
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
