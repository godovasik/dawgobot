package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/godovasik/dawgobot/internal/timeline"
)

// AddEvents массово добавляет события в базу данных
func (db *DB) AddEvents(events []timeline.Event) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO timeline (streamer_name, author, event_type, content, timestamp) 
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		_, err = stmt.Exec(
			event.Streamer,
			event.Author,
			int(event.Type),
			event.Content,
			event.Timestamp,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetEventsByTimeRange возвращает события стримера за указанный временной промежуток
func (db *DB) GetEventsByTimeRange(streamerName string, from, to time.Time) ([]timeline.Event, error) {
	query := `
		SELECT author, event_type, content, timestamp 
		FROM timeline 
		WHERE streamer_name = ? AND timestamp BETWEEN ? AND ? 
		ORDER BY timestamp ASC`

	rows, err := db.conn.Query(query, streamerName, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []timeline.Event
	for rows.Next() {
		var event timeline.Event
		var author sql.NullString
		var eventType int

		err := rows.Scan(&author, &eventType, &event.Content, &event.Timestamp)
		if err != nil {
			return nil, err
		}

		event.Streamer = streamerName
		event.Type = timeline.EventType(eventType)
		if author.Valid {
			event.Author = author.String
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// GetEventsByCount возвращает последние N событий стримера
func (db *DB) GetEventsByCount(streamerName string, count int) ([]timeline.Event, error) {
	query := `
		SELECT author, event_type, content, timestamp 
		FROM timeline 
		WHERE streamer_name = ? 
		ORDER BY timestamp DESC 
		LIMIT ?`

	rows, err := db.conn.Query(query, streamerName, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []timeline.Event
	for rows.Next() {
		var event timeline.Event
		var author sql.NullString
		var eventType int

		err := rows.Scan(&author, &eventType, &event.Content, &event.Timestamp)
		if err != nil {
			return nil, err
		}

		event.Streamer = streamerName
		event.Type = timeline.EventType(eventType)
		if author.Valid {
			event.Author = author.String
		}

		events = append(events, event)
	}

	// Разворачиваем массив, чтобы события шли в хронологическом порядке
	for i := range len(events) / 2 {
		events[i], events[len(events)-1-i] = events[len(events)-1-i], events[i]
	}

	return events, rows.Err()
}

// ExportEventsByTimeRangeToFile экспортирует события в текстовый файл
func (db *DB) ExportEventsByTimeRangeToFile(streamerName string, from, to time.Time) (string, error) {
	events, err := db.GetEventsByTimeRange(streamerName, from, to)
	if err != nil {
		return "", err
	}

	return db.exportEventsToFile(events, streamerName, from, to)
}

// ExportEventsByCountToFile экспортирует последние N событий в текстовый файл
func (db *DB) ExportEventsByCountToFile(streamerName string, count int) (string, error) {
	events, err := db.GetEventsByCount(streamerName, count)
	if err != nil {
		return "", err
	}

	var from, to time.Time
	if len(events) > 0 {
		from = events[0].Timestamp
		to = events[len(events)-1].Timestamp
	} else {
		now := time.Now()
		from = now
		to = now
	}

	return db.exportEventsToFile(events, streamerName, from, to)
}

// exportEventsToFile внутренняя функция для экспорта событий в файл
func (db *DB) exportEventsToFile(events []timeline.Event, streamerName string, from, to time.Time) (string, error) {
	// Создаем папку если не существует
	logsDir := "./logs/events"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return "", err
	}

	// Формируем имя файла: <ник>_<дата-начала>_<время-начала-время-конца>.log
	dateStr := from.Format("2006-01-02")
	timeStr := fmt.Sprintf("%s-%s", from.Format("15:04"), to.Format("15:04"))
	filename := fmt.Sprintf("%s_%s_%s.log", streamerName, dateStr, timeStr)
	filepath := filepath.Join(logsDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Записываем события в файл
	for _, event := range events {
		line := db.formatEventLine(event)
		if _, err := file.WriteString(line + "\n"); err != nil {
			return "", err
		}
	}

	return filepath, nil
}

// formatEventLine форматирует событие в строку для лог-файла
func (db *DB) formatEventLine(event timeline.Event) string {
	timeStr := event.Timestamp.Format("15:04:05")
	eventTypeStr := db.eventTypeToString(event.Type)

	switch event.Type {
	case timeline.EventChat:
		return fmt.Sprintf("[%s] [%s] %s: %s", timeStr, eventTypeStr, event.Author, event.Content)
	default:
		return fmt.Sprintf("[%s] [%s] %s", timeStr, eventTypeStr, event.Content)
	}
}

// eventTypeToString преобразует EventType в строку
func (db *DB) eventTypeToString(eventType timeline.EventType) string {
	switch eventType {
	case timeline.EventGlobal:
		return "GLOBAL"
	case timeline.EventChat:
		return "CHAT"
	case timeline.EventSpeech:
		return "SPEECH"
	case timeline.EventScreenshot:
		return "SCREENSHOT"
	default:
		return "UNKNOWN"
	}
}
