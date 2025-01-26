package middleware

import (
"log"
"os"
"path/filepath"
"time"

"fileditch/internal/db"
)

func StartCleanupJob() {
go func() {
ticker := time.NewTicker(15 * time.Minute)
defer ticker.Stop()

// Run cleanup immediately on start
if err := cleanup(); err != nil {
log.Printf("Initial cleanup error: %v", err)
}

for {
select {
case <-ticker.C:
if err := cleanup(); err != nil {
log.Printf("Cleanup error: %v", err)
}
}
}
}()

log.Println("Cleanup job started")
}

func cleanup() error {
log.Println("Running cleanup job...")

// Get expired files from database
expiredFiles, err := db.GetExpiredFiles()
if err != nil {
return err
}

if len(expiredFiles) == 0 {
log.Println("No expired files found")
return nil
}

log.Printf("Found %d expired files", len(expiredFiles))

for _, file := range expiredFiles {
// Delete the physical file
uploadDir := os.Getenv("UPLOAD_DIR")
if uploadDir == "" {
uploadDir = "uploads"
}
filePath := filepath.Join(uploadDir, file.Path)

if err := os.Remove(filePath); err != nil {
if !os.IsNotExist(err) {
log.Printf("Error deleting file %s: %v", filePath, err)
}
}

// Remove from database
if err := db.DeleteFile(file.ID); err != nil {
log.Printf("Error deleting file record %d: %v", file.ID, err)
continue
}

log.Printf("Deleted expired file: %s", file.Path)
}

// Remove empty subdirectories in uploads directory
if err := cleanEmptyDirs("uploads"); err != nil {
log.Printf("Error cleaning empty directories: %v", err)
}

return nil
}

func cleanEmptyDirs(dir string) error {
entries, err := os.ReadDir(dir)
if err != nil {
return err
}

// Recursively clean subdirectories
for _, entry := range entries {
if entry.IsDir() {
subdir := filepath.Join(dir, entry.Name())
if err := cleanEmptyDirs(subdir); err != nil {
return err
}
}
}

// Try to remove the directory if it's empty
entries, err = os.ReadDir(dir)
if err != nil {
return err
}

if len(entries) == 0 && dir != "uploads" { // Don't remove the root uploads directory
if err := os.Remove(dir); err != nil {
return err
}
log.Printf("Removed empty directory: %s", dir)
}

return nil
}