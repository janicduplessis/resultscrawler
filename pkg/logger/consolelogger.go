package logger

import (
	"fmt"
	"log"
)

// ConsoleLogger logs messages to standard output.
type ConsoleLogger struct{}

// Log logs a message.
func (handler *ConsoleLogger) Log(message string) {
	log.Println(message)
}

// Logf logs a message with format.
func (handler *ConsoleLogger) Logf(message string, args ...interface{}) {
	log.Println(fmt.Sprintf(message, args...))
}

// Error logs an error object.
func (handler *ConsoleLogger) Error(err error) {
	log.Println(err.Error())
}

// Fatal logs a fatal error. The application will panic.
func (handler *ConsoleLogger) Fatal(err error) {
	log.Fatal(err)
}
