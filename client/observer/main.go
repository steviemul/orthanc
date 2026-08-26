package main

import (
	"fmt"
	"orthanc-observer/event"
)

func main() {

	fmt.Println("orthanc-observer starting")

	var event = event.Event{
		EventType: "loaded",
		PID:       1234,
	}

	event.Update()

	eventJson, _ := event.Json()

	fmt.Printf("Processing event %s", eventJson)
}
