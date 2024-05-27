package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

var Conn *sql.DB

func InitDB() {
	var err error
	Conn, err = sql.Open("clickhouse", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err := Conn.Ping(); err != nil {
		log.Fatal(err)
	}

	_, err = Conn.Exec(`
	CREATE TABLE IF NOT EXISTS likes(
		post_id    INT NOT NULL,
		user_id    INT NOT NULL,
		PRIMARY KEY (post_id, user_id)
	) engine=ReplacingMergeTree`)
	if err != nil {
		log.Fatal("Error creating a table: ", err)
	}

	_, err = Conn.Exec(`
	CREATE TABLE IF NOT EXISTS views(
		post_id    INT NOT NULL,
		user_id    INT NOT NULL,
		PRIMARY KEY (post_id, user_id)
	) engine=ReplacingMergeTree`)
	if err != nil {
		log.Fatal("Error creating a table: ", err)
	}
}
