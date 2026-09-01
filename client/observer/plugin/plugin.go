package plugin

import (
	"github.com/steviemul/orthanc-observer/event"
)

type Plugin interface {
	CanHandle(e event.Event) bool
	ProcessEvent(e event.Event) *event.Evidence
}
