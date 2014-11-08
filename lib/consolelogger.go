package lib

import (
	"fmt"
	"log"
)

type ConsoleLogger struct{}

func (handler *ConsoleLogger) Log(message string) {
	log.Println(message)
}

func (handler *ConsoleLogger) Logf(message string, args ...interface{}) {
	log.Println(fmt.Sprintf(message, args))
}

func (handler *ConsoleLogger) Error(err error) {
	log.Println(err.Error())
}
