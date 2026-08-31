package main

import (
	"fmt"
	"time"

	"github.com/steviemul/orthanc-observer/event"
)

func main() {

	fmt.Println("orthanc-observer starting")

	pollingProcessViewer := event.NewPollingProcessViewer(time.Second, event.SystemOutEventHandler{})

	fmt.Println("orthanc-observer started. Ctrl+C to stop")

	pollingProcessViewer.Start()
}
