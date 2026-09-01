package main

import (
	"database/sql"
	"net/http"
	"strings"
)

type newsItem struct {
	ID, AuthorID                         int64
	Title, Content, CreatedAt, UpdatedAt string
}
type newsInput struct{ Title, Content string }

func scanNews(row interface{ Scan(...any) error }) (newsItem, error) {
	var value newsItem
	err := row.Scan(&value.ID, &value.AuthorID, &value.Title, &value.Content, &value.CreatedAt, &value.UpdatedAt)
	return value, err
}
func listNewsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id,author_id,title,content,created_at,updated_at FROM news ORDER BY created_at DESC")
		if err != nil {
			writeError(w, 500, "failed to load news")
			return
		}
		defer rows.Close()
		result := []newsItem{}
		for rows.Next() {
			value, err := scanNews(rows)
			if err != nil {
				writeError(w, 500, "failed to load news")
				return
			}
			result = append(result, value)
		}
		writeJSON(w, 200, result)
	}
}
func getNewsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		value, err := scanNews(db.QueryRow("SELECT id,author_id,title,content,created_at,updated_at FROM news WHERE id=?", id))
		if err != nil {
			writeError(w, 404, "news not found")
			return
		}
		writeJSON(w, 200, value)
	}
}
func createNewsHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input newsInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
			writeError(w, 400, "title and content are required")
			return
		}
		result, err := db.Exec("INSERT INTO news (author_id,title,content) VALUES (?,?,?)", authUser(r).ID, strings.TrimSpace(input.Title), strings.TrimSpace(input.Content))
		if err != nil {
			writeError(w, 500, "failed to create news")
			return
		}
		id, _ := result.LastInsertId()
		value, _ := scanNews(db.QueryRow("SELECT id,author_id,title,content,created_at,updated_at FROM news WHERE id=?", id))
		writeJSON(w, 201, value)
	})
}
func updateNewsHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		var input newsInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
			writeError(w, 400, "title and content are required")
			return
		}
		result, err := db.Exec("UPDATE news SET title=?,content=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", strings.TrimSpace(input.Title), strings.TrimSpace(input.Content), id)
		if err != nil || func() bool { n, _ := result.RowsAffected(); return n == 0 }() {
			writeError(w, 404, "news not found")
			return
		}
		value, _ := scanNews(db.QueryRow("SELECT id,author_id,title,content,created_at,updated_at FROM news WHERE id=?", id))
		writeJSON(w, 200, value)
	})
}
func deleteNewsHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		result, err := db.Exec("DELETE FROM news WHERE id=?", id)
		if err != nil {
			writeError(w, 500, "failed to delete news")
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			writeError(w, 404, "news not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
