package main

import (
	"database/sql"
	"net/http"
)

func listUsersHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id,email,name,role,created_at FROM users ORDER BY id")
		if err != nil {
			writeError(w, 500, "failed to load users")
			return
		}
		defer rows.Close()
		result := []user{}
		for rows.Next() {
			var value user
			if err := rows.Scan(&value.ID, &value.Email, &value.Name, &value.Role, &value.CreatedAt); err != nil {
				writeError(w, 500, "failed to load users")
				return
			}
			result = append(result, value)
		}
		writeJSON(w, 200, result)
	})
}
func updateUserRoleHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		var input struct {
			Role string `json:"role"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		if input.Role != "user" && input.Role != "moderator" && input.Role != "admin" {
			writeError(w, 400, "invalid role")
			return
		}
		result, err := db.Exec("UPDATE users SET role=? WHERE id=?", input.Role, id)
		if err != nil {
			writeError(w, 500, "failed to update role")
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			writeError(w, 404, "user not found")
			return
		}
		value, _ := getUser(db, id)
		writeJSON(w, 200, value)
	})
}
func deleteUserHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := idParam(r)
		if !ok {
			writeError(w, 400, "invalid id")
			return
		}
		if id == authUser(r).ID {
			writeError(w, 400, "cannot delete yourself")
			return
		}
		result, err := db.Exec("DELETE FROM users WHERE id=?", id)
		if err != nil {
			writeError(w, 500, "failed to delete user")
			return
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			writeError(w, 404, "user not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
