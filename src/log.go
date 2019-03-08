package main

import (
	"fmt"
	"log"
)

// LogLevel is the minimum level of log messages to print
type LogLevel int

// Auto set log level
const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
)

var level = INFO

// SetLogLevel sets minimum log level to print
func SetLogLevel(logLevel string) error {
	switch logLevel {
	case "FATAL":
		level = FATAL
	case "ERROR":
		level = ERROR
	case "WARNING":
		level = WARNING
	case "INFO":
		level = INFO
	case "DEBUG":
		level = DEBUG
	default:
		return fmt.Errorf("log level %s not allowed", logLevel)
	}
	return nil
}

// Debug prints very verbose messages
func Debug(message string, args ...interface{}) {
	if level <= DEBUG {
		log.Printf("DEBUG: "+message, args...)
	}
}

// Info prints informative messages
func Info(message string, args ...interface{}) {
	if level <= INFO {
		log.Printf("INFO: "+message, args...)
	}
}

// Warning prints messages you should give attention
func Warning(message string, args ...interface{}) {
	if level <= WARNING {
		log.Printf("WARNING: "+message, args...)
	}
}

// Error prints impacting messages
func Error(message string, args ...interface{}) {
	if level <= ERROR {
		log.Printf("ERROR: "+message, args...)
	}
}

// Fatal prints a message and exit program
func Fatal(message string, args ...interface{}) {
	if level <= FATAL {
		log.Fatalf("FATAL: "+message, args...)
	}
}
