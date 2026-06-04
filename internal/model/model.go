package model

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Summary   string    `json:"summary"`
	Author    string    `json:"author"`
	Status    string    `json:"status"` // draft, published
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Email    string `json:"email"`
}

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 创建表
	queries := []string{
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			summary TEXT,
			author TEXT DEFAULT 'admin',
			status TEXT DEFAULT 'draft',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			email TEXT
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return nil, err
		}
	}

	return db, nil
}
