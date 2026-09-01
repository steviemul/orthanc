package main

import (
	"fmt"
	"time"

	"github.com/steviemul/orthanc-observer/event"
)

func main() {

	fmt.Println("orthanc-observer starting")

	eventCh := make(chan event.Event)

	compositeEventHandler := event.NewCompositeEventHandler(
		event.SystemOutEventHandler{},
		event.FileOutEventHandler{},
	)

	go func() {
		for e := range eventCh {
			compositeEventHandler.HandleEvent(e)
		}
	}()

	pollingProcessViewer := event.NewPollingProcessViewer(
		time.Second,
		event.NewChannelEventHandler(eventCh),
	)

	fmt.Println("orthanc-observer started. Ctrl+C to stop")

	pollingProcessViewer.Start()
}
