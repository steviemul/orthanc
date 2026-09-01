package event

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type ProcessInfo struct {
	PID  int
	Path string
}

type PollingProcessViewer struct {
	interval time.Duration
	active   map[int]string
	handler  EventHandler
}

var _ ProcessViewer = (*PollingProcessViewer)(nil)

func NewPollingProcessViewer(interval time.Duration, handler EventHandler) *PollingProcessViewer {
	return &PollingProcessViewer{
		interval: interval,
		active:   make(map[int]string),
		handler:  handler,
	}
}

func (p *PollingProcessViewer) Snapshot() map[int]string {
	snap := make(map[int]string)

	entries, err := os.ReadDir("/proc")

	if err != nil {
		return snap
	}

	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())

		if err != nil {
			continue
		}

		exePath, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))

		if err != nil {
			continue
		}

		snap[pid] = exePath
	}

	return snap

}

func BuildEvent(eventType string, pid int, path string) Event {

	return Event{
		EventType: eventType,
		PID:       pid,
		Process:   path,
		Path:      path,
	}
}

func ProcessPids(eventType string, left map[int]string, right map[int]string, handler EventHandler) {

	for pid, path := range left {
		if _, exists := right[pid]; !exists {

			event := BuildEvent(eventType, pid, path)

			handler.HandleEvent(event)
		}
	}
}

func (p *PollingProcessViewer) Start() {
	p.active = p.Snapshot()

	ProcessPids("SNAPSHOT", p.active, make(map[int]string), p.handler)

	ticker := time.NewTicker(p.interval)

	defer ticker.Stop()

	for range ticker.C {
		current := p.Snapshot()

		ProcessPids("STARTED", current, p.active, p.handler)
		ProcessPids("STOPPED", p.active, current, p.handler)

		p.active = current
	}
}
