package routes

import (
	"html/template"
	"net/http"
	"path/filepath"
	"github.com/gorilla/sessions"
	"github.com/gorilla/mux"
	"log"
	"os"
)

var (
	store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))
)

func ServeLoginPage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		http.ServeFile(w, r, filepath.Join("app", "public", "upload.html"))
	} else {
		http.ServeFile(w, r, filepath.Join("app", "public", "login.html"))
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	password := r.FormValue("password")
	if password == os.Getenv("SITE_PASSWORD") {
		session.Values["authenticated"] = true
		session.Save(r, w)
		w.Write([]byte(`{"success": true}`))
	} else {
		log.Printf("Failed login attempt from IP: %s", r.RemoteAddr)
		http.Error(w, `{"error": "Invalid password"}`, http.StatusUnauthorized)
	}
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	w.Write([]byte(`{"success": true}`))
}

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/", ServeLoginPage).Methods("GET")
	r.HandleFunc("/login", HandleLogin).Methods("POST")
	r.HandleFunc("/logout", HandleLogout).Methods("POST")
}
