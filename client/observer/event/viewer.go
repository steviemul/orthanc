package event

// ProcessViewer watches for processes starting and stopping, emitting an
// Event for each to its EventHandler. PollingProcessViewer and
// EBPFProcessViewer are interchangeable implementations.
type ProcessViewer interface {
	Start()
}
