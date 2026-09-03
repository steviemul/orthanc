package main

import (
	"fmt"
	"time"

	"github.com/steviemul/orthanc-observer/config"
	"github.com/steviemul/orthanc-observer/event"
	"github.com/steviemul/orthanc-observer/plugin"
)

func main() {

	fmt.Println("orthanc-observer starting")

	eventCh := make(chan event.Event)

	compositeEventHandler := event.NewCompositeEventHandler(
		event.SystemOutEventHandler{},
		event.FileOutEventHandler{},
		event.NewRestHandler(config.GetEventCollectorHost()),
	)

	go func() {
		for e := range eventCh {
			e.Evidence = plugin.RunPipeline(e)

			compositeEventHandler.HandleEvent(e)
		}
	}()

	channelHandler := event.NewChannelEventHandler(eventCh)

	var processViewer event.ProcessViewer

	if config.UseEBPF() {
		processViewer = event.NewEBPFProcessViewer(channelHandler)
	} else {
		processViewer = event.NewPollingProcessViewer(time.Second, channelHandler)
	}

	fmt.Println("orthanc-observer started. Ctrl+C to stop")

	processViewer.Start()
}
