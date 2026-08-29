package main

import (
	"fmt"
	"orthanc-observer/event"
	"time"
)

func main() {

	fmt.Println("orthanc-observer starting")

	pollingProcessViewer := event.NewPollingProcessViewer(2*time.Second, event.SystemOutEventHandler{})

	fmt.Println("orthanc-observer started. Ctrl+C to stop")

	pollingProcessViewer.Start()
}
