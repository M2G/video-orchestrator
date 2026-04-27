package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(connString string) *pgxpool.Pool {

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatal("unable to connect to database:", err)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("database unreachable:", err)
	}

	log.Println("PostgreSQL connected (pgx)")

	return pool
}
