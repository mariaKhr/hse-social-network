package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func InitDB() {
	var err error
	Pool, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to create connection pool:", err)
	}
}

func CloseDB() {
	Pool.Close()
}

func CreateTable() {
	_, err := Pool.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS posts(
		post_id    INT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
		user_id    INT NOT NULL,
		content    TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL);
	`)

	if err != nil {
		log.Fatal("Error creating a table:", err)
	}
}