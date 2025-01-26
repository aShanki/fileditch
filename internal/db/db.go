package db

import (
"database/sql"
"os"
"path/filepath"
"time"

_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type File struct {
ID        int64
Name      string
Path      string
ExpiresAt time.Time
CreatedAt time.Time
}

// Initialize creates a new database connection and sets up the schema
func Initialize() error {
dbPath := os.Getenv("DB_PATH")
if dbPath == "" {
dbPath = "data/fileditch.db"
}

// Ensure directory exists
if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
return err
}

var err error
db, err = sql.Open("sqlite3", dbPath)
if err != nil {
return err
}

// Create tables
_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS files (
id INTEGER PRIMARY KEY AUTOINCREMENT,
name TEXT NOT NULL,
path TEXT NOT NULL UNIQUE,
expires_at DATETIME NOT NULL,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`)
return err
}

// Close closes the database connection
func Close() error {
if db != nil {
return db.Close()
}
return nil
}

// SaveFile saves a new file record to the database
func SaveFile(name, path string, expiresAt time.Time) error {
_, err := db.Exec(
"INSERT INTO files (name, path, expires_at) VALUES (?, ?, ?)",
name, path, expiresAt,
)
return err
}

// GetFile retrieves a file record by its path
func GetFile(path string) (*File, error) {
var file File
err := db.QueryRow(
"SELECT id, name, path, expires_at, created_at FROM files WHERE path = ?",
path,
).Scan(&file.ID, &file.Name, &file.Path, &file.ExpiresAt, &file.CreatedAt)

if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}
return &file, nil
}

// GetExpiredFiles returns all files that have expired
func GetExpiredFiles() ([]File, error) {
rows, err := db.Query(
"SELECT id, name, path, expires_at, created_at FROM files WHERE expires_at <= CURRENT_TIMESTAMP",
)
if err != nil {
return nil, err
}
defer rows.Close()

var files []File
for rows.Next() {
var file File
err := rows.Scan(&file.ID, &file.Name, &file.Path, &file.ExpiresAt, &file.CreatedAt)
if err != nil {
return nil, err
}
files = append(files, file)
}
return files, rows.Err()
}

// DeleteFile removes a file record from the database
func DeleteFile(id int64) error {
_, err := db.Exec("DELETE FROM files WHERE id = ?", id)
return err
}

// DeleteExpiredFiles removes all expired files from the database
func DeleteExpiredFiles() error {
_, err := db.Exec("DELETE FROM files WHERE expires_at <= CURRENT_TIMESTAMP")
return err
}