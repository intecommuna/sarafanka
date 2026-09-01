package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ad struct {
	ID, UserID, Price                                                    int64
	Title, Description, Category, ImageURL, Status, CreatedAt, UpdatedAt string
}
type adInput struct {
	Title, Description, Category, ImageURL, Status string
	Price                                          int64
}

func scanAd(row interface{ Scan(...any) error }) (ad, error) {
	var value ad
	err := row.Scan(&value.ID, &value.UserID, &value.Title, &value.Description, &value.Price, &value.Category, &value.ImageURL, &value.Status, &value.CreatedAt, &value.UpdatedAt)
	return value, err
}
func listAdsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT id,user_id,title,description,price,category,image_url,status,created_at,updated_at FROM ads`
		args := []any{}
		conditions := []string{}
		role := ""
		if value := r.Header.Get("Authorization"); value != "" {
			parts := strings.Fields(value)
			if len(parts) == 2 {
				if token, err := jwtUser(parts[1], loadConfig().JWTSecret); err == nil {
					role = token.Role
				}
			}
		}
		if role != "admin" && role != "moderator" {
			conditions = append(conditions, "status = 'active'")
		}
		if category := r.URL.Query().Get("category"); category != "" {
			conditions = append(conditions, "category = ?")
			args = append(args, category)
		}
		if search := r.URL.Query().Get("q"); search != "" {
			conditions = append(conditions, "(title LIKE ? OR description LIKE ?)")
			value := "%" + search + "%"
			args = append(args, value, value)
		}
		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}
		query += " ORDER BY created_at DESC"
		rows, err := db.Query(query, args...)
		if err != nil {
			writeError(w, 500, "failed to load ads")
			return
		}
		defer rows.Close()
		result := []ad{}
		for rows.Next() {
			value, err := scanAd(rows)
			if err != nil {
				writeError(w, 500, "failed to load ads")
				return
			}
			result = append(result, value)
		}
		writeJSON(w, 200, result)
	}
}
func getAdHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		value, err := scanAd(db.QueryRow("SELECT id,user_id,title,description,price,category,image_url,status,created_at,updated_at FROM ads WHERE id = ?", id))
		if err != nil {
			writeError(w, 404, "ad not found")
			return
		}
		if value.Status != "active" {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 {
				writeError(w, 404, "ad not found")
				return
			}
			claims, err := jwtUser(parts[1], loadConfig().JWTSecret)
			if err != nil || (claims.Role != "moderator" && claims.Role != "admin") {
				writeError(w, 404, "ad not found")
				return
			}
		}
		writeJSON(w, 200, value)
	}
}
func createAdHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input adInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Description) == "" || input.Price < 0 {
			writeError(w, 400, "title and description are required; price must be non-negative")
			return
		}
		if input.Category == "" {
			input.Category = "other"
		}
		result, err := db.Exec("INSERT INTO ads (user_id,title,description,price,category,image_url,status) VALUES (?, ?, ?, ?, ?, ?, ?)", authUser(r).ID, strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), input.Price, input.Category, input.ImageURL, valueOr(input.Status, "active"))
		if err != nil {
			writeError(w, 500, "failed to create ad")
			return
		}
		id, _ := result.LastInsertId()
		value, _ := scanAd(db.QueryRow("SELECT id,user_id,title,description,price,category,image_url,status,created_at,updated_at FROM ads WHERE id = ?", id))
		writeJSON(w, 201, value)
	})
}
func updateAdHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		var owner int64
		if err := db.QueryRow("SELECT user_id FROM ads WHERE id = ?", id).Scan(&owner); err != nil {
			writeError(w, 404, "ad not found")
			return
		}
		if !canModifyAd(r, owner) {
			writeError(w, 403, "forbidden")
			return
		}
		var input adInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Description) == "" || input.Price < 0 {
			writeError(w, 400, "title and description are required; price must be non-negative")
			return
		}
		_, err := db.Exec("UPDATE ads SET title=?,description=?,price=?,category=?,image_url=?,status=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), input.Price, valueOr(input.Category, "other"), input.ImageURL, valueOr(input.Status, "active"), id)
		if err != nil {
			writeError(w, 500, "failed to update ad")
			return
		}
		value, _ := scanAd(db.QueryRow("SELECT id,user_id,title,description,price,category,image_url,status,created_at,updated_at FROM ads WHERE id = ?", id))
		writeJSON(w, 200, value)
	})
}
func deleteAdHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		var owner int64
		if err := db.QueryRow("SELECT user_id FROM ads WHERE id=?", id).Scan(&owner); err != nil {
			writeError(w, 404, "ad not found")
			return
		}
		if !canModifyAd(r, owner) {
			writeError(w, 403, "forbidden")
			return
		}
		if _, err := db.Exec("DELETE FROM ads WHERE id=?", id); err != nil {
			writeError(w, 500, "failed to delete ad")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
func jwtUser(raw, secret string) (authClaims, error) {
	token, err := parseToken(raw, secret)
	if err != nil {
		return authClaims{}, err
	}
	return token, nil
}
func parseToken(raw, secret string) (authClaims, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		return authClaims{}, strconv.ErrSyntax
	}
	claims := token.Claims.(jwt.MapClaims)
	uid, ok := claims["uid"].(float64)
	if !ok {
		return authClaims{}, strconv.ErrSyntax
	}
	role, _ := claims["role"].(string)
	return authClaims{ID: int64(uid), Role: role}, nil
}
