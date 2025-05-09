// Copyright (C) 2025 Juan Antonio Gomez Pena
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Package logging provides application-wide logging capabilities
// with support for file-based logs, rotating logs, and different
// log levels based on environment.
package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/newbpydev/tusk/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logger references
var (
	// Logger is the global logger instance
	Logger *zap.Logger

	// Access these through helper functions below
	consoleLogger *zap.Logger
	fileLogger    *zap.Logger

	// Dedicated file-only logger (no console output)
	FileOnlyLogger *zap.Logger

	// Flag to indicate if we're in quiet mode (no console output)
	quietMode bool
)

// Component specific loggers for different parts of the application
var (
	CLILogger     *zap.Logger
	DBLogger      *zap.Logger
	ServiceLogger *zap.Logger
	APILogger     *zap.Logger
	TUILogger     *zap.Logger
)

// LogLevel represents the different logging levels
type LogLevel string

// LogLevel constants
const (
	DebugLevel   LogLevel = "debug"
	InfoLevel    LogLevel = "info"
	WarningLevel LogLevel = "warning"
	ErrorLevel   LogLevel = "error"
)

// Init initializes the logging system with the given configuration.
// It sets up both console and file logging based on environment.
func Init(cfg *config.Config) error {
	return InitWithOptions(cfg, false)
}

// InitWithOptions initializes the logging system with the given configuration and options.
// When quietMode is true, all logs are directed to the log file only, not to the console.
func InitWithOptions(cfg *config.Config, quiet bool) error {
	// Store the quiet mode flag
	quietMode = quiet

	// Create logs directory if it doesn't exist
	logDir := getLogDirectory()

	// Remove the debug print statements that display on screen
	// and just create the directory silently
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Set up console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// Set up file encoder - use JSON for machine readability
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.TimeKey = "timestamp"
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	// Determine the log level based on environment
	logLevel := getLogLevelFromEnvironment(cfg.AppEnv)
	zapLevel := getZapLogLevel(logLevel)

	// Set up console core - only show warnings and errors in production
	consoleLevel := zapLevel
	if cfg.AppEnv == "production" {
		consoleLevel = zapcore.WarnLevel // Only warnings and above in production
	}

	// In quiet mode, set console level to a level that effectively disables console output
	if quietMode {
		consoleLevel = zapcore.FatalLevel + 1 // Higher than any defined level to suppress all output
	}

	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.Lock(os.Stderr),
		consoleLevel,
	)

	// Set up file core with rotation
	logFilePath := filepath.Join(logDir, "tusk.log")

	// Try to create the file explicitly to verify we have write permissions
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Continue with console logging only
		Logger = zap.New(consoleCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		// Close the file immediately - we just wanted to verify permissions
		file.Close()

		// Configure lumberjack for log rotation
		fileWriter := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    10, // megabytes
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		}

		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(fileWriter),
			zapLevel,
		)

		// In quiet mode, only use the file core for all loggers
		if quietMode {
			Logger = zap.New(fileCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
			FileOnlyLogger = Logger // In quiet mode, main logger is already file-only
		} else {
			// Normal mode - combine cores for standard logger
			core := zapcore.NewTee(consoleCore, fileCore)
			Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

			// Create a dedicated file-only logger
			FileOnlyLogger = zap.New(fileCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		}
	}

	// Create component specific loggers
	CLILogger = Logger.Named("cli")
	DBLogger = Logger.Named("db")
	ServiceLogger = Logger.Named("service")
	APILogger = Logger.Named("api")

	// TUI logger always uses the file-only logger to avoid console output
	if FileOnlyLogger != nil {
		TUILogger = FileOnlyLogger.Named("tui")
	} else {
		TUILogger = Logger.Named("tui")
	}

	// Log initialization message (will go to the file only in quiet mode)
	logMessage := "Logging system initialized"
	if quietMode {
		logMessage += " in quiet mode (console output disabled)"
	}
	Logger.Info(logMessage)

	return nil
}

// getLogLevelFromEnvironment returns the appropriate log level for the environment.
func getLogLevelFromEnvironment(env string) LogLevel {
	switch env {
	case "production":
		return InfoLevel
	case "staging":
		return DebugLevel
	default: // development or any other
		return DebugLevel
	}
}

// getZapLogLevel converts our LogLevel to zapcore.Level
func getZapLogLevel(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarningLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// getLogDirectory returns the directory where log files will be stored.
// It checks for LOG_DIR environment variable, otherwise uses a default location.
// This function is enhanced to work with Bash in VS Code on Windows.
func getLogDirectory() string {
	// First check if LOG_DIR is specified
	if dir, exists := os.LookupEnv("LOG_DIR"); exists && dir != "" {
		// Make sure the path is using OS-specific separators
		return filepath.FromSlash(dir)
	}

	// For Windows, use %APPDATA%\Tusk\logs (more compatible with Bash)
	// For Unix/Linux, use ~/.tusk/logs
	var baseDir string

	// Check for APPDATA which is more reliable than LOCALAPPDATA in Bash on Windows
	if appData := os.Getenv("APPDATA"); appData != "" {
		// Windows with Bash
		baseDir = filepath.Join(appData, "Tusk")
	} else if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		// Windows with regular shell
		baseDir = filepath.Join(localAppData, "Tusk")
	} else {
		// Unix/Linux/macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// Fallback to current directory if we can't get home directory
			return "logs"
		}
		baseDir = filepath.Join(homeDir, ".tusk")
	}

	return filepath.Join(baseDir, "logs")
}

// Sync flushes any buffered log entries.
func Sync() error {
	return Logger.Sync()
}
