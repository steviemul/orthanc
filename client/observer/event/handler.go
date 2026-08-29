package event

import (
	"fmt"
	"os"
)

const LOG_FILE = "/var/log/observer.log"

type EventHandler interface {
	HandleEvent(e Event)
}

type SystemOutEventHandler struct{}

func (eh SystemOutEventHandler) HandleEvent(e Event) {
	e.Update()

	eventJson, _ := e.Json()

	fmt.Printf("Processing event %s\n", eventJson)

	fileHandler := FileOutEventHandler{}

	fileHandler.HandleEvent(e)
}

type FileOutEventHandler struct{}

func (fh FileOutEventHandler) HandleEvent(e Event) {
	e.Update()

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err == nil {
		eventJson, _ := e.Json()

		logFile.Write(eventJson)
		logFile.Write([]byte("\n"))

		logFile.Close()
	}
}
