package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	initDb()
}

func initDb() {
	createFilesTable()
	createAccessLogsTable()
}

func createFilesTable() {
	query := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		url_path TEXT NOT NULL UNIQUE,
		mime_type TEXT,
		size INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		downloads INTEGER DEFAULT 0
	)`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating files table: %v", err)
	}
}

func createAccessLogsTable() {
	query := `
	CREATE TABLE IF NOT EXISTS access_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(file_id) REFERENCES files(id)
	)`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Error creating access_logs table: %v", err)
	}
}
