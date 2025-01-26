package routes

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var db *sql.DB

func InitFileRoutes(router *mux.Router, database *sql.DB) {
	db = database
	router.HandleFunc("/file/{filename}", serveFile).Methods("GET")
	router.HandleFunc("/file/{filename}/info", getFileInfo).Methods("GET")
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	var file File
	err := db.QueryRow(`
		SELECT * FROM files 
		WHERE url_path = ? 
		AND expires_at > datetime('now')`,
		filename).Scan(&file.ID, &file.OriginalName, &file.FilePath, &file.URLPath, &file.MimeType, &file.Size, &file.CreatedAt, &file.ExpiresAt, &file.Downloads)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			logrus.Error("Database error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	// Log access
	_, err = db.Exec(`
		INSERT INTO access_logs (file_id, ip_address, user_agent)
		VALUES (?, ?, ?)`,
		file.ID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		logrus.Error("Failed to log access:", err)
	}

	// Update download count
	_, err = db.Exec(`
		UPDATE files SET downloads = downloads + 1 WHERE id = ?`,
		file.ID)
	if err != nil {
		logrus.Error("Failed to update download count:", err)
	}

	// Set content disposition and type
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Disposition", `inline; filename="`+file.OriginalName+`"`)

	// Handle client disconnection
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		logrus.Info("Client disconnected during file transfer")
	}()

	// Send file with absolute path and error handling
	http.ServeFile(w, r, filepath.Clean(file.FilePath))
}

func getFileInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	var file File
	err := db.QueryRow(`
		SELECT original_name, mime_type, size, downloads, created_at, expires_at 
		FROM files 
		WHERE url_path = ? 
		AND expires_at > datetime('now')`,
		filename).Scan(&file.OriginalName, &file.MimeType, &file.Size, &file.Downloads, &file.CreatedAt, &file.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			logrus.Error("Database error:", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}

type File struct {
	ID           int
	OriginalName string
	FilePath     string
	URLPath      string
	MimeType     string
	Size         int64
	CreatedAt    time.Time
	ExpiresAt    time.Time
	Downloads    int
}
