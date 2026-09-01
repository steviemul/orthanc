package event

import (
	"fmt"
	"os"

	"github.com/steviemul/orthanc-observer/config"
)

type EventHandler interface {
	HandleEvent(e Event)
}

type CompositeEventHandler struct {
	handlers []EventHandler
}

func (eh CompositeEventHandler) HandleEvent(e Event) {

	for _, handler := range eh.handlers {
		handler.HandleEvent(e)
	}
}

func NewCompositeEventHandler(handlers ...EventHandler) CompositeEventHandler {
	return CompositeEventHandler{
		handlers: handlers,
	}
}

type ChannelEventHandler struct {
	ch chan Event
}

func (h ChannelEventHandler) HandleEvent(e Event) {
	h.ch <- e
}

func NewChannelEventHandler(ch chan Event) ChannelEventHandler {
	return ChannelEventHandler{
		ch: ch,
	}
}

type SystemOutEventHandler struct{}

func (eh SystemOutEventHandler) HandleEvent(e Event) {
	e.Update()

	eventJson, _ := e.Json()

	fmt.Printf("Processing event %s\n", eventJson)
}

type FileOutEventHandler struct{}

func (fh FileOutEventHandler) HandleEvent(e Event) {
	e.Update()

	logFile, err := os.OpenFile(config.LOG_FILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	eventJson, _ := e.Json()

	logFile.Write(eventJson)
	logFile.Write([]byte("\n"))
}
