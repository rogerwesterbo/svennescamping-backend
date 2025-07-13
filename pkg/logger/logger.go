package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	Logger *zap.Logger
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"

	// Background colors
	RedBg = "\033[41m"
)

// coloredWriter wraps the output to add colors to the entire line
type coloredWriter struct{}

func newColoredWriter() *coloredWriter {
	return &coloredWriter{}
}

func (cw *coloredWriter) Write(p []byte) (n int, err error) {
	line := string(p)

	// Determine color based on log level in the line
	var color string
	if strings.Contains(line, "DEBUG") {
		color = Purple
	} else if strings.Contains(line, "INFO") {
		color = Green
	} else if strings.Contains(line, "WARN") {
		color = Yellow
	} else if strings.Contains(line, "ERROR") {
		color = Red
	} else if strings.Contains(line, "FATAL") {
		color = RedBg + White
	} else if strings.Contains(line, "PANIC") {
		color = RedBg + White
	} else {
		color = Reset
	}

	// Apply color to entire line
	coloredLine := color + strings.TrimRight(line, "\n") + Reset + "\n"
	return os.Stdout.Write([]byte(coloredLine))
}

// InitLogger initializes the global logger based on environment with colored output
func InitLogger(development bool) error {
	var config zapcore.EncoderConfig
	var logLevel zapcore.Level

	if development {
		config = zap.NewDevelopmentEncoderConfig()
		logLevel = zapcore.DebugLevel
	} else {
		config = zap.NewProductionEncoderConfig()
		logLevel = zapcore.InfoLevel
	}

	// Configure encoder for console output
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeDuration = zapcore.SecondsDurationEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder // Use capital level names without color

	// Create encoder and core with custom colored writer
	encoder := zapcore.NewConsoleEncoder(config)
	writer := zapcore.AddSync(newColoredWriter())
	core := zapcore.NewCore(encoder, writer, logLevel)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Fallback to a default logger if not initialized
		Logger, _ = zap.NewProduction()
	}
	return Logger
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// HTTPRequestStart logs the start of an HTTP request
func HTTPRequestStart(method, path, remoteAddr, userAgent string) {
	Info("HTTP request started",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("remote_addr", remoteAddr),
		zap.String("user_agent", userAgent),
	)
}

// HTTPRequestComplete logs the completion of an HTTP request
func HTTPRequestComplete(method, path string, duration time.Duration, statusCode int) {
	Info("HTTP request completed",
		zap.String("method", method),
		zap.String("path", path),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	)
}
