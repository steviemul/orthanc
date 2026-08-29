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

func (p *PollingProcessViewer) Start() {
	p.active = p.Snapshot()

	ticker := time.NewTicker(p.interval)

	defer ticker.Stop()

	for range ticker.C {
		current := p.Snapshot()

		for pid, path := range current {
			if _, exists := p.active[pid]; !exists {

				event := BuildEvent("started", pid, path)

				p.handler.HandleEvent(event)
			}
		}

		for pid, path := range p.active {
			if _, exists := current[pid]; !exists {
				event := BuildEvent("stopped", pid, path)

				p.handler.HandleEvent(event)
			}
		}

		p.active = current
	}
}
