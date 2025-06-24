package database

import (
	"database/sql"

	"github.com/godovasik/dawgobot/logger"
	_ "github.com/mattn/go-sqlite3"
)

type EventType int

// const (
// 	EventGlobal EventType = iota
// 	EventChat
// 	EventSpeech
// 	EventScreenshot
// )

// type Event struct {
// 	Type      EventType
// 	Content   string
// 	Author    string
// 	Streamer  string
// 	Timestamp time.Time
// }

type DB struct {
	conn *sql.DB
}

func New() (*DB, error) {
	dbPath := "internal/database/db.sqlite"
	conn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		conn.Close()
		return nil, err
	}
	logger.Info("db initialized")

	return db, nil
}

func (db *DB) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS timeline (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		streamer_name TEXT NOT NULL,
		author TEXT,
		event_type INTEGER NOT NULL,
		content TEXT NOT NULL,
		timestamp DATETIME NOT NULL
	);

	-- Составной индекс для быстрых запросов по стримеру и времени
	CREATE INDEX IF NOT EXISTS idx_timeline_streamer_time 
	ON timeline(streamer_name, timestamp);

	-- Дополнительный индекс по типу событий для фильтрации
	CREATE INDEX IF NOT EXISTS idx_timeline_event_type 
	ON timeline(event_type);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}
