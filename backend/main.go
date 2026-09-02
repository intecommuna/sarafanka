package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

type healthResponse struct {
	Status string `json:"status"`
}

func loadDotEnv() {
	for _, path := range []string{".env", "backend/.env"} {
		file, err := os.Open(path)
		if err != nil || file == nil {
			continue
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "" || os.Getenv(key) != "" {
				continue
			}
			_ = os.Setenv(key, value)
		}
	}
}

func main() {
	loadDotEnv()
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	config := loadConfig()
	if err := seedDB(db, config); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})
	mux.HandleFunc("POST /api/auth/register", registerHandler(db, config))
	mux.HandleFunc("POST /api/auth/login", loginHandler(db, config))
	mux.Handle("GET /api/auth/me", requireAuth(config, http.HandlerFunc(meHandler(db))))
	mux.HandleFunc("GET /api/ads", listAdsHandler(db))
	mux.HandleFunc("GET /api/ads/{id}", getAdHandler(db))
	mux.Handle("POST /api/ads", requireAuth(config, createAdHandler(db)))
	mux.Handle("PUT /api/ads/{id}", requireAuth(config, updateAdHandler(db)))
	mux.Handle("DELETE /api/ads/{id}", requireAuth(config, deleteAdHandler(db)))
	mux.HandleFunc("GET /api/news", listNewsHandler(db))
	mux.HandleFunc("GET /api/news/{id}", getNewsHandler(db))
	mux.Handle("POST /api/news", requireRole(config, "moderator", "admin", createNewsHandler(db)))
	mux.Handle("PUT /api/news/{id}", requireRole(config, "moderator", "admin", updateNewsHandler(db)))
	mux.Handle("DELETE /api/news/{id}", requireRole(config, "moderator", "admin", deleteNewsHandler(db)))
	mux.Handle("GET /api/admin/users", requireRole(config, "admin", listUsersHandler(db)))
	mux.Handle("PUT /api/admin/users/{id}/role", requireRole(config, "admin", updateUserRoleHandler(db)))
	mux.Handle("DELETE /api/admin/users/{id}", requireRole(config, "admin", deleteUserHandler(db)))
	mux.Handle("GET /api/tools/parser", requireAuth(config, http.HandlerFunc(parserHandler)))
	mux.Handle("GET /api/tools/analytics", requireAuth(config, http.HandlerFunc(analyticsHandler)))
	mux.Handle("GET /api/tools/mortgage", requireAuth(config, http.HandlerFunc(mortgageHandler)))

	log.Println("Backend server running on :8080")
	if err := http.ListenAndServe(":8080", cors(mux)); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
