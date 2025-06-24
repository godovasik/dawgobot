package logger

import (
	"log"
	"os"
	"strings"
)

// ANSI цветовые коды
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[37m"

	// Жирные версии
	ColorBoldRed    = "\033[1;31m"
	ColorBoldYellow = "\033[1;33m"
	ColorBoldBlue   = "\033[1;34m"
	ColorBoldGray   = "\033[1;37m"
)

// LogLevel определяет уровень логгирования
type LogLevel int

const (
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
)

// Logger структура для логгера
type Logger struct {
	level     LogLevel
	logger    *log.Logger
	useColors bool
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
	useColors := shouldUseColors()

	return &Logger{
		level:     level,
		logger:    log.New(os.Stdout, "", log.LstdFlags),
		useColors: useColors,
	}
}

// shouldUseColors определяет, нужно ли использовать цвета
func shouldUseColors() bool {
	// Отключаем цвета если:
	// 1. Явно отключено через переменную окружения
	// 2. Вывод не в терминал (например, в файл или pipe)
	// 3. CI/CD среда

	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Проверяем, что stdout это терминал
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	return true
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
	case "DEBUG":
		return DEBUG
	default:
		return INFO // уровень по умолчанию
	}
}

// colorize окрашивает текст если цвета включены
func (l *Logger) colorize(color, text string) string {
	if !l.useColors {
		return text
	}
	return color + text + ColorReset
}

// getLevelPrefix возвращает цветной префикс для уровня
func (l *Logger) getLevelPrefix(level LogLevel) string {
	switch level {
	case ERROR:
		return l.colorize(ColorBoldRed, "[ERROR]")
	case WARN:
		return l.colorize(ColorBoldYellow, "[WARN]")
	case INFO:
		return l.colorize(ColorBoldBlue, "[INFO]")
	case DEBUG:
		return l.colorize(ColorBoldGray, "[DEBUG]")
	default:
		return "[UNKNOWN]"
	}
}

// Error выводит сообщение уровня ERROR
func (l *Logger) Error(msg string) {
	if l.level >= ERROR {
		l.logger.Printf("%s %s", l.getLevelPrefix(ERROR), msg)
	}
}

// Warn выводит сообщение уровня WARN
func (l *Logger) Warn(msg string) {
	if l.level >= WARN {
		l.logger.Printf("%s %s", l.getLevelPrefix(WARN), msg)
	}
}

// Info выводит сообщение уровня INFO
func (l *Logger) Info(msg string) {
	if l.level >= INFO {
		l.logger.Printf("%s %s", l.getLevelPrefix(INFO), msg)
	}
}

// Debug выводит сообщение уровня DEBUG
func (l *Logger) Debug(msg string) {
	if l.level >= DEBUG {
		l.logger.Printf("%s %s", l.getLevelPrefix(DEBUG), msg)
	}
}

// Errorf выводит форматированное сообщение уровня ERROR
func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.level >= ERROR {
		l.logger.Printf("%s "+format, append([]interface{}{l.getLevelPrefix(ERROR)}, args...)...)
	}
}

// Warnf выводит форматированное сообщение уровня WARN
func (l *Logger) Warnf(format string, args ...interface{}) {
	if l.level >= WARN {
		l.logger.Printf("%s "+format, append([]interface{}{l.getLevelPrefix(WARN)}, args...)...)
	}
}

// Infof выводит форматированное сообщение уровня INFO
func (l *Logger) Infof(format string, args ...interface{}) {
	if l.level >= INFO {
		l.logger.Printf("%s "+format, append([]interface{}{l.getLevelPrefix(INFO)}, args...)...)
	}
}

// Debugf выводит форматированное сообщение уровня DEBUG
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.level >= DEBUG {
		l.logger.Printf("%s "+format, append([]interface{}{l.getLevelPrefix(DEBUG)}, args...)...)
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

func Debug(msg string) {
	globalLogger.Debug(msg)
}

func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// GetLogger возвращает глобальный логгер для прямого использования
func GetLogger() *Logger {
	return globalLogger
}

// SetLogger устанавливает новый глобальный логгер
func SetLogger(l *Logger) {
	globalLogger = l
}
