package main

import (
	"database/sql"
	"log"
	"time"
	"fmt"
	"filescanner"
)

type Foto struct {
	Filename string
}

func Count(db *sql.DB, query string, args ...interface{}) (int) {
	rows, err := db.Query(query, args ...)
	if err != nil {
		log.Fatal("Failed to get count: ", err)
	}
	defer rows.Close()

	rows.Next()
	var count int
	rows.Scan(&count)
	return count
}

type SqlPopulator struct {
	db *sql.DB
}

func (sp SqlPopulator) visitImageFile(path string, modTime time.Time) {
	c := Count(sp.db, "SELECT COUNT(1) FROM fotos WHERE path = ?", path)
	if c > 0 {
		return
	}

	sp.db.Exec("INSERT INTO fotos (path, mtime) VALUES (?, ?)", path, modTime)
	fmt.Println("Added imageFile: ", path)
}

func createTable(db *sql.DB) {
	rows, err := db.Query("SELECT COUNT(*) FROM sqlite_master WHERE type = ? AND name = ?", "table", "fotos")
	if err != nil {
		log.Fatal(err)
	}

	rows.Next()
	var count int
	rows.Scan(&count)
	rows.Close()

	if count < 1 {
		_, err := db.Exec(`
			CREATE TABLE fotos (id INTEGER NOT NULL PRIMARY KEY, path TEXT NOT NULL, mtime DATETIME, rotation INTEGER)
		`)
		if err != nil {
			log.Fatal("Fail to create table: ", err)
		}
		fmt.Println("Created table fotos.")
	}
}

func fillSqlLiteDb() {
	db, err := sql.Open("sqlite3", "./fotos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)

	sp := SqlPopulator{db}
	filescanner.Scan("/mnt/nas/Pictures/boon-phone-sync/2017", sp.visitImageFile)
}