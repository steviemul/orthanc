package main

import (
	"fmt"
	"orthanc-observer/event"
	"time"
)

func main() {

	fmt.Println("orthanc-observer starting")

	var process = event.Process

	process(
		event.DummyObserver{},
		event.SystemOutEventHandler{},
		10,
		time.Second,
	)
}
