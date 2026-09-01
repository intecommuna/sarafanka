package main

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strconv"
	"strings"
)

type authClaims struct {
	ID          int64
	Email, Role string
}
type contextKey string

const authKey contextKey = "auth"

func authUser(r *http.Request) authClaims { return r.Context().Value(authKey).(authClaims) }
func requireAuth(config appConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, 401, "authorization required")
			return
		}
		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, 401, "invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, 401, "invalid token")
			return
		}
		uid, ok := claims["uid"].(float64)
		if !ok {
			writeError(w, 401, "invalid token")
			return
		}
		role, _ := claims["role"].(string)
		email, _ := claims["email"].(string)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authKey, authClaims{int64(uid), email, role})))
	})
}
func requireRole(config appConfig, roles ...any) http.Handler {
	next := roles[len(roles)-1].(http.Handler)
	allowed := map[string]bool{}
	for _, role := range roles[:len(roles)-1] {
		allowed[role.(string)] = true
	}
	return requireAuth(config, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowed[authUser(r).Role] {
			writeError(w, 403, "forbidden")
			return
		}
		next.ServeHTTP(w, r)
	}))
}
func canModifyAd(r *http.Request, ownerID int64) bool {
	u := authUser(r)
	return u.ID == ownerID || u.Role == "moderator" || u.Role == "admin"
}
func idParam(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
