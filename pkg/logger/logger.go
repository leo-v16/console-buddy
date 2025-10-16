package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents different log levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging with different levels
type Logger struct {
	level      LogLevel
	logger     *log.Logger
	logFile    *os.File
	enableFile bool
}

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Output     io.Writer
	LogFile    string
	EnableFile bool
	Prefix     string
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      INFO,
		Output:     os.Stdout,
		EnableFile: false,
		Prefix:     "[Console-AI] ",
	}
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	logger := &Logger{
		level:      config.Level,
		enableFile: config.EnableFile,
	}

	var writers []io.Writer
	if config.Output != nil {
		writers = append(writers, config.Output)
	}

	// Setup file logging if enabled
	if config.EnableFile && config.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		logger.logFile = file
		writers = append(writers, file)
	}

	// Create multi-writer if we have multiple outputs
	var output io.Writer = os.Stdout
	if len(writers) > 0 {
		if len(writers) == 1 {
			output = writers[0]
		} else {
			output = io.MultiWriter(writers...)
		}
	}

	logger.logger = log.New(output, config.Prefix, 0)

	return logger, nil
}

// Close closes the logger and any open file handles
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// shouldLog determines if a message should be logged based on the current level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// formatMessage formats a log message with timestamp, level, and caller information
func (l *Logger) formatMessage(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Get caller information
	_, file, line, ok := runtime.Caller(3) // Skip formatMessage, log method, and public method
	var caller string
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	} else {
		caller = "unknown"
	}

	return fmt.Sprintf("%s [%s] %s - %s", timestamp, level.String(), caller, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.shouldLog(DEBUG) {
		message := fmt.Sprintf(format, args...)
		l.logger.Println(l.formatMessage(DEBUG, message))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.shouldLog(INFO) {
		message := fmt.Sprintf(format, args...)
		l.logger.Println(l.formatMessage(INFO, message))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.shouldLog(WARN) {
		message := fmt.Sprintf(format, args...)
		l.logger.Println(l.formatMessage(WARN, message))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.shouldLog(ERROR) {
		message := fmt.Sprintf(format, args...)
		l.logger.Println(l.formatMessage(ERROR, message))
	}
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	if l.shouldLog(FATAL) {
		message := fmt.Sprintf(format, args...)
		l.logger.Println(l.formatMessage(FATAL, message))
		l.Close()
		os.Exit(1)
	}
}

// ErrorWithStack logs an error message with stack trace
func (l *Logger) ErrorWithStack(err error, format string, args ...interface{}) {
	if l.shouldLog(ERROR) {
		message := fmt.Sprintf(format, args...)
		if err != nil {
			message = fmt.Sprintf("%s: %v", message, err)
		}

		// Add stack trace
		buf := make([]byte, 1024)
		for {
			n := runtime.Stack(buf, false)
			if n < len(buf) {
				buf = buf[:n]
				break
			}
			buf = make([]byte, 2*len(buf))
		}

		fullMessage := fmt.Sprintf("%s\nStack trace:\n%s", message, string(buf))
		l.logger.Println(l.formatMessage(ERROR, fullMessage))
	}
}

// LogToolCall logs a tool call with its parameters
func (l *Logger) LogToolCall(toolName string, params map[string]interface{}) {
	if l.shouldLog(DEBUG) {
		message := fmt.Sprintf("\nTool call: %s with params: %+v", toolName, params)
		l.logger.Println(l.formatMessage(DEBUG, message))
	}
}

// LogToolResult logs a tool call result
func (l *Logger) LogToolResult(toolName string, success bool, result interface{}, err error) {
	level := INFO
	if !success {
		level = ERROR
	}

	if l.shouldLog(level) {
		var message string
		if success {
			message = fmt.Sprintf("\nTool %s completed successfully: %+v", toolName, result)
		} else {
			message = fmt.Sprintf("\nTool %s failed: %v", toolName, err)
		}
		l.logger.Println(l.formatMessage(level, message))
	}
}

// LogConversation logs conversation messages
func (l *Logger) LogConversation(role, message string) {
	if l.shouldLog(DEBUG) {
		// Truncate very long messages for logging
		truncated := message
		if len(message) > 500 {
			truncated = message[:500] + "..."
		}
		logMessage := fmt.Sprintf("\nConversation [%s]: %s", role, truncated)
		l.logger.Println(l.formatMessage(DEBUG, logMessage))
	}
}

// Performance logging
type PerformanceTimer struct {
	logger    *Logger
	operation string
	startTime time.Time
}

// StartTimer starts a performance timer for the given operation
func (l *Logger) StartTimer(operation string) *PerformanceTimer {
	if l.shouldLog(DEBUG) {
		l.Debug("\nStarting operation: %s", operation)
	}
	return &PerformanceTimer{
		logger:    l,
		operation: operation,
		startTime: time.Now(),
	}
}

// Stop stops the performance timer and logs the duration
func (pt *PerformanceTimer) Stop() {
	duration := time.Since(pt.startTime)
	if pt.logger.shouldLog(DEBUG) {
		pt.logger.Debug("Operation %s completed in %v", pt.operation, duration)
	}
}

// Global logger instance
var defaultLogger *Logger

// Initialize sets up the default logger
func Initialize(config *Config) error {
	var err error
	defaultLogger, err = NewLogger(config)
	return err
}

// Shutdown closes the default logger
func Shutdown() {
	if defaultLogger != nil {
		defaultLogger.Close()
	}
}

// Global logging functions using the default logger
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Fatal(format, args...)
	}
}

func ErrorWithStack(err error, format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.ErrorWithStack(err, format, args...)
	}
}

func LogToolCall(toolName string, params map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.LogToolCall(toolName, params)
	}
}

func LogToolResult(toolName string, success bool, result interface{}, err error) {
	if defaultLogger != nil {
		defaultLogger.LogToolResult(toolName, success, result, err)
	}
}

func LogConversation(role, message string) {
	if defaultLogger != nil {
		defaultLogger.LogConversation(role, message)
	}
}

func StartTimer(operation string) *PerformanceTimer {
	if defaultLogger != nil {
		return defaultLogger.StartTimer(operation)
	}
	return &PerformanceTimer{
		operation: operation,
		startTime: time.Now(),
	}
}
