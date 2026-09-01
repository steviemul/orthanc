package plugin

import (
	"github.com/steviemul/orthanc-observer/event"
)

var plugins = []Plugin{
	BusyboxPlugin{},
}

func RunPipeline(e event.Event) []event.Evidence {

	results := make([]event.Evidence, 0)

	for _, plugin := range plugins {
		if plugin.CanHandle(e) {
			evidence := plugin.ProcessEvent(e)

			if evidence != nil {
				results = append(results, *evidence)
			}
		}
	}

	return results
}
