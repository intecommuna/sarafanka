package main

import (
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type user struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}
type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}
func registerHandler(db *sql.DB, config appConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input credentials
		if !decodeJSON(w, r, &input) {
			return
		}
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		input.Name = strings.TrimSpace(input.Name)
		if input.Email == "" || input.Password == "" || input.Name == "" {
			writeError(w, 400, "email, password and name are required")
			return
		}
		hash, err := hashPassword(input.Password)
		if err != nil {
			writeError(w, 500, "failed to hash password")
			return
		}
		result, err := db.Exec("INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)", input.Email, hash, input.Name)
		if err != nil {
			writeError(w, 409, "email is already registered")
			return
		}
		id, _ := result.LastInsertId()
		current, err := getUser(db, id)
		if err != nil {
			writeError(w, 500, "failed to load user")
			return
		}
		token, err := createToken(current, config.JWTSecret)
		if err != nil {
			writeError(w, 500, "failed to create token")
			return
		}
		writeJSON(w, 201, map[string]any{"token": token, "user": current})
	}
}
func loginHandler(db *sql.DB, config appConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input credentials
		if !decodeJSON(w, r, &input) {
			return
		}
		var id int64
		var hash string
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", strings.ToLower(strings.TrimSpace(input.Email))).Scan(&id, &hash)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
			writeError(w, 401, "invalid email or password")
			return
		}
		current, err := getUser(db, id)
		if err != nil {
			writeError(w, 500, "failed to load user")
			return
		}
		token, err := createToken(current, config.JWTSecret)
		if err != nil {
			writeError(w, 500, "failed to create token")
			return
		}
		writeJSON(w, 200, map[string]any{"token": token, "user": current})
	}
}
func meHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, err := getUser(db, authUser(r).ID)
		if err != nil {
			writeError(w, 404, "user not found")
			return
		}
		writeJSON(w, 200, current)
	}
}
func getUser(db *sql.DB, id int64) (user, error) {
	var u user
	err := db.QueryRow("SELECT id, email, name, role, created_at FROM users WHERE id = ?", id).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt)
	return u, err
}
func createToken(u user, secret string) (string, error) {
	claims := jwt.MapClaims{"uid": u.ID, "email": u.Email, "role": u.Role, "exp": time.Now().Add(7 * 24 * time.Hour).Unix()}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, 400, "invalid JSON")
		return false
	}
	return true
}
