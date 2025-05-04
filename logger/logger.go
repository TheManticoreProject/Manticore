package logger

import (
	"fmt"
	"sync"
	"time"
)

var LoggerLock sync.Mutex

func Lock() {
	LoggerLock.Lock()
}

func Unlock() {
	LoggerLock.Unlock()
}

// INFO

func Info(message string) {
	DatePrintf("INFO: %s\n", message)
}

func InfoMilliseconds(message string) {
	DatePrintfMilliseconds("INFO: %s\n", message)
}

func InfoMicroseconds(message string) {
	DatePrintfMicroseconds("INFO: %s\n", message)
}

func InfoNanoseconds(message string) {
	DatePrintfNanoseconds("INFO: %s\n", message)
}

// WARN

func Warn(message string) {
	DatePrintf("WARN: %s\n", message)
}

func WarnMilliseconds(message string) {
	DatePrintfMilliseconds("WARN: %s\n", message)
}

func WarnMicroseconds(message string) {
	DatePrintfMicroseconds("WARN: %s\n", message)
}

func WarnNanoseconds(message string) {
	DatePrintfNanoseconds("WARN: %s\n", message)
}

// ERROR

func Error(message string) {
	DatePrintf("ERROR: %s\n", message)
}

func ErrorMilliseconds(message string) {
	DatePrintfMilliseconds("ERROR: %s\n", message)
}

func ErrorMicroseconds(message string) {
	DatePrintfMicroseconds("ERROR: %s\n", message)
}

func ErrorNanoseconds(message string) {
	DatePrintfNanoseconds("ERROR: %s\n", message)
}

// DEBUG

func Debug(message string) {
	DatePrintfMilliseconds("DEBUG: %s\n", message)
}

func DebugMilliseconds(message string) {
	DatePrintfMilliseconds("DEBUG: %s\n", message)
}

func DebugMicroseconds(message string) {
	DatePrintfMicroseconds("DEBUG: %s\n", message)
}

func DebugNanoseconds(message string) {
	DatePrintfNanoseconds("DEBUG: %s\n", message)
}

// PRINT

func Print(message string) {
	DatePrintf("%s\n", message)
}

func PrintMilliseconds(message string) {
	DatePrintfMilliseconds("%s\n", message)
}

func PrintMicroseconds(message string) {
	DatePrintfMicroseconds("%s\n", message)
}

func PrintNanoseconds(message string) {
	DatePrintfNanoseconds("%s\n", message)
}

// DatePrintf

func DatePrintf(format string, message ...any) {
	currentTime := time.Now().Format("2006-01-02 15h04m05s")
	format = fmt.Sprintf("[%s] %s", currentTime, format)
	fmt.Printf(format, message...)
}

func DatePrintfMilliseconds(format string, message ...any) {
	currentTime := time.Now().Format("2006-01-02 15h04m05s.0000")
	format = fmt.Sprintf("[%s] %s", currentTime, format)
	fmt.Printf(format, message...)
}

func DatePrintfMicroseconds(format string, message ...any) {
	currentTime := time.Now().Format("2006-01-02 15h04m05s.000000")
	format = fmt.Sprintf("[%s] %s", currentTime, format)
	fmt.Printf(format, message...)
}

func DatePrintfNanoseconds(format string, message ...any) {
	currentTime := time.Now().Format("2006-01-02 15h04m05s.000000000")
	format = fmt.Sprintf("[%s] %s", currentTime, format)
	fmt.Printf(format, message...)
}
