package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type appConfig struct{ JWTSecret, AdminEmail, AdminPassword string }

func loadConfig() appConfig {
	return appConfig{getEnv("JWT_SECRET", "dev-secret-change-me"), getEnv("ADMIN_EMAIL", "admin@sarafanka.su"), getEnv("ADMIN_PASSWORD", "admin123")}
}
func getEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
func dbPath() string {
	if value := os.Getenv("DB_PATH"); value != "" {
		return value
	}
	if _, err := os.Stat("/app"); err == nil {
		return "/app/data/sarafanka.db"
	}
	return "./data/sarafanka.db"
}
func openDB() (*sql.DB, error) {
	path := dbPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
func migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, name TEXT NOT NULL, role TEXT NOT NULL DEFAULT 'user', created_at TEXT DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS ads (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, title TEXT NOT NULL, description TEXT NOT NULL, price INTEGER NOT NULL DEFAULT 0, category TEXT NOT NULL DEFAULT 'other', image_url TEXT DEFAULT '', status TEXT DEFAULT 'active', created_at TEXT DEFAULT CURRENT_TIMESTAMP, updated_at TEXT DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS news (id INTEGER PRIMARY KEY AUTOINCREMENT, author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, title TEXT NOT NULL, content TEXT NOT NULL, created_at TEXT DEFAULT CURRENT_TIMESTAMP, updated_at TEXT DEFAULT CURRENT_TIMESTAMP);`)
	if err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	rows, err := db.Query("PRAGMA table_info(ads)")
	if err != nil {
		return fmt.Errorf("inspect ads table: %w", err)
	}
	defer rows.Close()
	hasType := false
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == "type" {
			hasType = true
		}
	}
	if !hasType {
		if _, err := db.Exec("ALTER TABLE ads ADD COLUMN type TEXT NOT NULL DEFAULT 'product'"); err != nil {
			return fmt.Errorf("add ads type: %w", err)
		}
	}
	return nil
}
func seedDB(db *sql.DB, config appConfig) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return seedServices(db)
	}
	hash, err := hashPassword(config.AdminPassword)
	if err != nil {
		return err
	}
	result, err := db.Exec("INSERT INTO users (email, password_hash, name, role) VALUES (?, ?, ?, 'admin')", config.AdminEmail, hash, "Администратор")
	if err != nil {
		return err
	}
	adminID, _ := result.LastInsertId()
	ads := []struct {
		title, description, category string
		price                        int
	}{{"Велосипед", "Городской велосипед в хорошем состоянии", "other", 25000}, {"Ноутбук", "Рабочий ноутбук для учебы и дома", "electronics", 50000}, {"Квартира", "Уютная квартира рядом с метро", "real-estate", 7500000}}
	for _, ad := range ads {
		if _, err := db.Exec("INSERT INTO ads (user_id, title, description, price, category, type) VALUES (?, ?, ?, ?, ?, 'product')", adminID, ad.title, ad.description, ad.price, ad.category); err != nil {
			return err
		}
	}
	for _, item := range []struct{ title, content string }{{"Добро пожаловать", "Добро пожаловать на Sarafanka!"}, {"Правила площадки", "Будьте вежливы и публикуйте достоверные объявления."}} {
		if _, err := db.Exec("INSERT INTO news (author_id, title, content) VALUES (?, ?, ?)", adminID, item.title, item.content); err != nil {
			return err
		}
	}
	return nil
}

func seedServices(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM ads WHERE type = 'service'").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	var adminID int64
	if err := db.QueryRow("SELECT id FROM users WHERE role = 'admin' ORDER BY id LIMIT 1").Scan(&adminID); err != nil {
		return err
	}
	services := []struct{ title, description, category string }{{"Ремонт квартир под ключ", "Полный комплекс ремонтных работ для вашей квартиры.", "repair"}, {"Репетитор по математике", "Индивидуальные занятия по математике для школьников.", "education"}, {"Грузоперевозки по городу", "Аккуратные городские грузоперевозки на автомобиле.", "transport_services"}}
	for _, service := range services {
		if _, err := db.Exec("INSERT INTO ads (user_id,title,description,price,category,type) VALUES (?,?,?,?,?,'service')", adminID, service.title, service.description, 0, service.category); err != nil {
			return err
		}
	}
	return nil
}
