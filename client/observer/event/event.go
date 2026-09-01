package event

import (
	"encoding/json"
	"time"
)

type Evidence struct {
	Facts map[string]string `json:"facts"`
}

type Event struct {
	Timestamp time.Time  `json:"timestamp"`
	EventType string     `json:"event_type"`
	PID       int        `json:"pid"`
	Process   string     `json:"process"`
	Path      string     `json:"path"`
	Source    string     `json:"source"`
	Evidence  []Evidence `json:"evidence"`
}

func (e *Event) Update() {
	e.Timestamp = time.Now()
}

func (e *Event) Json() ([]byte, error) {
	return json.Marshal(e)
}
