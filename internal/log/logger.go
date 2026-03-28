package log

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// LogLevel represents log severity
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Logger provides structured logging
type Logger struct {
	level      LogLevel
	jsonOutput bool
}

var defaultLogger *Logger

// init creates the default logger
func init() {
	defaultLogger = NewLogger(InfoLevel)
}

// NewLogger creates a new logger
func NewLogger(level LogLevel) *Logger {
	return &Logger{
		level:      level,
		jsonOutput: os.Getenv("LOG_FORMAT") == "json",
	}
}

// SetJSONOutput enables JSON logging
func (l *Logger) SetJSONOutput(enabled bool) {
	l.jsonOutput = enabled
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(DebugLevel, msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.log(WarnLevel, msg, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.log(ErrorLevel, msg, fields...)
}

// log outputs a log entry
func (l *Logger) log(level LogLevel, msg string, fields ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	if l.jsonOutput {
		l.logJSON(level, msg, fields...)
	} else {
		l.logText(level, msg, fields...)
	}
}

// shouldLog returns true if message should be logged at this level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DebugLevel: 0,
		InfoLevel:  1,
		WarnLevel:  2,
		ErrorLevel: 3,
	}

	return levels[level] >= levels[l.level]
}

// logText outputs text-formatted log
func (l *Logger) logText(level LogLevel, msg string, fields ...interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s: %s", timestamp, level, msg)

	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}

	fmt.Println()
}

// logJSON outputs JSON-formatted log
func (l *Logger) logJSON(level LogLevel, msg string, fields ...interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     string(level),
		Message:   msg,
		Fields:    make(map[string]interface{}),
	}

	// Parse fields (key-value pairs)
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			entry.Fields[key] = fields[i+1]
		}
	}

	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}

// Global functions
func Debug(msg string, fields ...interface{}) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	defaultLogger.Error(msg, fields...)
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func SetJSONOutput(enabled bool) {
	defaultLogger.SetJSONOutput(enabled)
}
