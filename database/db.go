package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() *sql.DB {

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URI"))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Successfully connected to Postgresql Server !")
	}
	return db
}
