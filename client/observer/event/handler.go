package event

import "fmt"

type EventHandler interface {
	HandleEvent(e Event)
}

type SystemOutEventHandler struct{}

func (eh SystemOutEventHandler) HandleEvent(e Event) {
	e.Update()

	eventJson, _ := e.Json()

	fmt.Printf("Processing event %s\n", eventJson)
}
