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

package logging

import (
	"go.uber.org/zap"
)

// GetComponentLogger returns a named logger for a specific component.
// This allows for finer-grained logging in different parts of the application.
func GetComponentLogger(component string) *zap.Logger {
	if Logger == nil {
		// Safety check - should not happen in normal operation
		panic("Logger not initialized. Call logging.Init() first")
	}
	return Logger.Named(component)
}

// Debug logs a message at DebugLevel with associated structured context.
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info logs a message at InfoLevel with associated structured context.
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn logs a message at WarnLevel with associated structured context.
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error logs a message at ErrorLevel with associated structured context.
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal logs a message at FatalLevel with associated structured context,
// then exits with status code 1.
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}

// GetLoggerForContext returns the appropriate logger based on the context.
// This is useful for middleware or shared code that might be used by different components.
func GetLoggerForContext(ctx string) *zap.Logger {
	switch ctx {
	case "cli":
		return CLILogger
	case "db":
		return DBLogger
	case "service":
		return ServiceLogger
	case "api":
		return APILogger
	case "tui":
		return TUILogger
	default:
		return Logger
	}
}

// WithFields adds structured context to a logger.
func WithFields(logger *zap.Logger, fields map[string]interface{}) *zap.Logger {
	if logger == nil {
		return nil
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return logger.With(zapFields...)
}
