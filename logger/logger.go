package logger

import (
	"log"
	"os"
	"strings"
)

// LogLevel определяет уровень логгирования
type LogLevel int

const (
	ERROR LogLevel = iota
	WARN
	INFO
)

// Logger структура для логгера
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// Глобальный экземпляр логгера
var globalLogger *Logger

// init инициализирует глобальный логгер при загрузке пакета
func init() {
	globalLogger = NewLogger()
}

// NewLogger создает новый логгер с уровнем из переменной окружения LOG_LEVEL
func NewLogger() *Logger {
	level := getLogLevelFromEnv()
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// getLogLevelFromEnv получает уровень логгирования из переменной окружения
func getLogLevelFromEnv() LogLevel {
	envLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch envLevel {
	case "ERROR":
		return ERROR
	case "WARN":
		return WARN
	case "INFO":
		return INFO
	default:
		return INFO // уровень по умолчанию
	}
}

// Error выводит сообщение уровня ERROR
func (l *Logger) Error(msg string) {
	if l.level >= ERROR {
		l.logger.Printf("[ERROR] %s", msg)
	}
}

// Warn выводит сообщение уровня WARN
func (l *Logger) Warn(msg string) {
	if l.level >= WARN {
		l.logger.Printf("[WARN] %s", msg)
	}
}

// Info выводит сообщение уровня INFO
func (l *Logger) Info(msg string) {
	if l.level >= INFO {
		l.logger.Printf("[INFO] %s", msg)
	}
}

// Глобальные функции для удобного использования
func Error(msg string) {
	globalLogger.Error(msg)
}

func Warn(msg string) {
	globalLogger.Warn(msg)
}

func Info(msg string) {
	globalLogger.Info(msg)
}

// Errorf выводит форматированное сообщение уровня ERROR
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level >= ERROR {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

// Warnf выводит форматированное сообщение уровня WARN
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.level >= WARN {
		l.logger.Printf("[WARN] "+format, args...)
	}
}

// Infof выводит форматированное сообщение уровня INFO
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level >= INFO {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

// Глобальные функции с форматированием для удобного использования
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// GetLogger возвращает глобальный логгер для прямого использования
func GetLogger() *Logger {
	return globalLogger
}

// SetLogger устанавливает новый глобальный логгер
func SetLogger(l *Logger) {
	globalLogger = l
}
