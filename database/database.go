package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Init() *sql.DB {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=docker password=docker dbname=shortlinks sslmode=disable")

	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected!")
	return db

}

func Migrate(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tb_links (
			id SERIAL PRIMARY KEY, 
			long_url TEXT NOT NULL, 
			short_url TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Migrated!")
}
