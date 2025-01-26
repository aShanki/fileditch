package middleware

import (
	"net/http"
	"os"
	"log"
)

// Authentication middleware
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth, ok := r.Context().Value("isAuthenticated").(bool); ok && auth {
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}

// Password verification
func VerifyPassword(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")

	if password == os.Getenv("SITE_PASSWORD") {
		ctx := context.WithValue(r.Context(), "isAuthenticated", true)
		r = r.WithContext(ctx)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	} else {
		// Log failed attempt
		log.Printf("Failed login attempt from IP: %s", r.RemoteAddr)
		http.Error(w, `{"error": "Invalid password"}`, http.StatusUnauthorized)
	}
}
