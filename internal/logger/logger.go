package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles application logging
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	logger   *log.Logger
	level    Level
	filePath string
	maxSize  int64 // Max file size in bytes
}

var (
	instance *Logger
	once     sync.Once
)

// Init initializes the global logger
func Init(logDir string, level Level) error {
	var err error
	once.Do(func() {
		instance, err = newLogger(logDir, level)
	})
	return err
}

// newLogger creates a new logger instance
func newLogger(logDir string, level Level) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFileName := fmt.Sprintf("netvision_%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(logDir, logFileName)

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)

	l := &Logger{
		file:     file,
		logger:   log.New(multiWriter, "", 0),
		level:    level,
		filePath: logPath,
		maxSize:  10 * 1024 * 1024, // 10 MB
	}

	return l, nil
}

// Get returns the global logger instance
func Get() *Logger {
	if instance == nil {
		// Create a default console-only logger if not initialized
		instance = &Logger{
			logger: log.New(os.Stdout, "", 0),
			level:  LevelInfo,
		}
	}
	return instance
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log writes a log entry
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if we need to rotate the log file
	l.checkRotation()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), message)

	l.logger.Println(logLine)
}

// checkRotation checks if log file needs rotation
func (l *Logger) checkRotation() {
	if l.file == nil {
		return
	}

	info, err := l.file.Stat()
	if err != nil {
		return
	}

	if info.Size() >= l.maxSize {
		l.rotate()
	}
}

// rotate rotates the log file
func (l *Logger) rotate() {
	if l.file == nil {
		return
	}

	l.file.Close()

	// Rename current file with timestamp
	rotatedPath := l.filePath + "." + time.Now().Format("150405")
	os.Rename(l.filePath, rotatedPath)

	// Open new file
	file, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.file = nil
		l.logger = log.New(os.Stdout, "", 0)
		return
	}

	l.file = file
	multiWriter := io.MultiWriter(os.Stdout, file)
	l.logger = log.New(multiWriter, "", 0)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetFilePath returns the current log file path
func (l *Logger) GetFilePath() string {
	return l.filePath
}

// Package-level convenience functions

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	Get().Debug(format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	Get().Info(format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	Get().Warn(format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	Get().Error(format, args...)
}

// CleanOldLogs removes log files older than specified days
func CleanOldLogs(logDir string, maxAgeDays int) error {
	if maxAgeDays <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)

	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			return nil
		}

		// Only delete .log files
		if filepath.Ext(path) != ".log" {
			return nil
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(path)
		}

		return nil
	})
}
