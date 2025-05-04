// Package utils provides utility functions for the Bubble Tea TUI
package utils

import (
	"fmt"
	"os"
	"reflect"
	"time"
)

// DebugLog logs debug information to a file
func DebugLog(format string, args ...interface{}) {
	// Open debug log file in append mode
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	// Format the log with timestamp
	timestamp := time.Now().Format("15:04:05.000")
	logMsg := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
	
	// Write to file
	f.WriteString(logMsg)
}

// LogMsg logs information about a message
func LogMsg(prefix string, msg interface{}) {
	DebugLog("%s: Type=%s Value=%+v", prefix, reflect.TypeOf(msg), msg)
}
