package event

import "math/rand"

var EVENT_TYPES = [...]string{
	"LOADED",
	"STARTED",
	"FINISHED",
}

var PROCESSES = [...]string{
	"grep",
	"cat",
	"java",
}

func get_pid() int {
	return rand.Intn(10_000)
}

func get_event_type() string {
	return EVENT_TYPES[rand.Intn(len(EVENT_TYPES))]
}

func get_process() string {
	return PROCESSES[rand.Intn(len(PROCESSES))]
}

type Observer interface {
	GetEvent() Event
}

type DummyObserver struct{}

func (o DummyObserver) GetEvent() Event {
	process := get_process()

	return Event{
		EventType: get_event_type(),
		PID:       get_pid(),
		Process:   process,
		Path:      "/usr/bin/" + process,
	}
}
