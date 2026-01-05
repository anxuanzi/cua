// Package logging provides structured logging for CUA.
package logging

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	// LevelDebug is for verbose debugging information.
	LevelDebug Level = iota
	// LevelInfo is for general informational messages.
	LevelInfo
	// LevelWarn is for warning messages.
	LevelWarn
	// LevelError is for error messages.
	LevelError
	// LevelNone disables all logging.
	LevelNone
)

// String returns the string representation of a log level.
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
		return "NONE"
	}
}

// Logger is the main logging interface.
type Logger struct {
	mu       sync.Mutex
	level    Level
	output   io.Writer
	prefix   string
	useColor bool
}

// Colors for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorCyan   = "\033[36m"
)

// defaultLogger is the package-level logger.
var defaultLogger = New(LevelInfo, os.Stderr)

// New creates a new Logger.
func New(level Level, output io.Writer) *Logger {
	return &Logger{
		level:    level,
		output:   output,
		useColor: isTerminal(output),
	}
}

// isTerminal checks if the output is a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		// Simple heuristic: check if it's stdout or stderr
		return f == os.Stdout || f == os.Stderr
	}
	return false
}

// SetLevel sets the log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
	l.useColor = isTerminal(w)
}

// SetPrefix sets the log prefix.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// WithPrefix returns a new logger with the given prefix.
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:    l.level,
		output:   l.output,
		prefix:   prefix,
		useColor: l.useColor,
	}
}

// log writes a log message.
func (l *Logger) log(level Level, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	now := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	var levelStr string
	var color string
	if l.useColor {
		switch level {
		case LevelDebug:
			color = colorGray
		case LevelInfo:
			color = colorBlue
		case LevelWarn:
			color = colorYellow
		case LevelError:
			color = colorRed
		}
		levelStr = fmt.Sprintf("%s%-5s%s", color, level.String(), colorReset)
	} else {
		levelStr = fmt.Sprintf("%-5s", level.String())
	}

	prefix := ""
	if l.prefix != "" {
		if l.useColor {
			prefix = fmt.Sprintf("%s[%s]%s ", colorCyan, l.prefix, colorReset)
		} else {
			prefix = fmt.Sprintf("[%s] ", l.prefix)
		}
	}

	fmt.Fprintf(l.output, "%s %s %s%s\n", now, levelStr, prefix, msg)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...any) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...any) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...any) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...any) {
	l.log(LevelError, format, args...)
}

// Package-level functions that use the default logger.

// SetLevel sets the default logger's level.
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// SetOutput sets the default logger's output.
func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// Debug logs a debug message.
func Debug(format string, args ...any) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message.
func Info(format string, args ...any) {
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...any) {
	defaultLogger.Warn(format, args...)
}

// Error logs an error message.
func Error(format string, args ...any) {
	defaultLogger.Error(format, args...)
}

// WithPrefix returns a new logger with the given prefix.
func WithPrefix(prefix string) *Logger {
	return defaultLogger.WithPrefix(prefix)
}

// ParseLevel parses a log level string.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "none", "off":
		return LevelNone
	default:
		return LevelInfo
	}
}

// ToolLogger is a specialized logger for tool operations.
type ToolLogger struct {
	*Logger
	toolName string
}

// NewToolLogger creates a logger for a specific tool.
func NewToolLogger(toolName string) *ToolLogger {
	return &ToolLogger{
		Logger:   defaultLogger.WithPrefix(toolName),
		toolName: toolName,
	}
}

// Start logs the start of a tool operation.
func (t *ToolLogger) Start(operation string, args ...any) {
	argsStr := formatArgs(args)
	t.Debug("→ %s(%s)", operation, argsStr)
}

// Success logs a successful tool operation.
func (t *ToolLogger) Success(operation string, result any) {
	t.Debug("✓ %s → %v", operation, result)
}

// Failure logs a failed tool operation.
func (t *ToolLogger) Failure(operation string, err error) {
	t.Error("✗ %s → %v", operation, err)
}

// formatArgs formats arguments for logging.
func formatArgs(args []any) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}
	return strings.Join(parts, ", ")
}

// Platform logs platform information at startup.
func LogPlatformInfo() {
	Info("Platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}
