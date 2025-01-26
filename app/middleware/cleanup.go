package middleware

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
}

func cleanupExpiredFiles() {
	rows, err := db.Query(`SELECT id, file_path FROM files WHERE expires_at < datetime('now')`)
	if err != nil {
		log.Printf("Error querying expired files: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var filePath string
		if err := rows.Scan(&id, &filePath); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		if err := os.Remove(filePath); err != nil {
			log.Printf("Error deleting file %s: %v", filePath, err)
			continue
		}

		if _, err := db.Exec(`DELETE FROM files WHERE id = ?`, id); err != nil {
			log.Printf("Error deleting file record %d: %v", id, err)
		}

		if _, err := db.Exec(`DELETE FROM access_logs WHERE file_id = ?`, id); err != nil {
			log.Printf("Error deleting access log for file %d: %v", id, err)
		}

		log.Printf("Cleaned up expired file: %s", filePath)
	}
}

func startCleanupJob() {
	cleanupExpiredFiles()

	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			cleanupExpiredFiles()
		}
	}()
}
