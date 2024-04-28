package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
)

var Conn *sql.DB

func InitDB() {
	var err error
	Conn, err = sql.Open("clickhouse", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err := Conn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
	}

	_, err = Conn.Exec(`
	CREATE TABLE IF NOT EXISTS likes(
		post_id    INT NOT NULL,
		user_id    INT NOT NULL
	) engine=Memory`)
	if err != nil {
		log.Fatal("Error creating a table: ", err)
	}

	_, err = Conn.Exec(`
	CREATE TABLE IF NOT EXISTS views(
		post_id    INT NOT NULL,
		user_id    INT NOT NULL
	) engine=Memory`)
	if err != nil {
		log.Fatal("Error creating a table: ", err)
	}
}
