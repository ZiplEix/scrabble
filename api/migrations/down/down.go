package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	dsn := flag.String("dsn", "", "Postgres DSN, example: postgres://user:pwd@host:port/dbname?sslmode=disable")
	dir := flag.String("dir", "..", "répertoire des migrations")
	flag.Parse()

	if *dsn == "" {
		fmt.Println("DSN is required")
		flag.Usage()
		return
	}

	if *dir == "" {
		fmt.Println("Migration directory is required")
		flag.Usage()
		return
	}

	var db *sql.DB
	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_ = goose.SetDialect("postgres")

	if err := goose.Down(db, *dir); err != nil {
		panic(err)
	}

	if err := goose.Run("status", db, *dir); err != nil {
		panic(err)
	}
}
