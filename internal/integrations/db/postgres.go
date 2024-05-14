package db

import (
	"context"
	"log"

	"github.com/go-pg/pg/v10"
)

func InitPostgres(host string, database string, user string, password string) *pg.DB {
	db := pg.Connect(&pg.Options{
		Addr:     host,
		User:     user,
		Password: password,
		Database: database,
	})
	_, err := db.ExecContext(context.Background(), "SELECT 1")

	if err != nil {
		log.Printf("%v", err)
	}

	return db
}
