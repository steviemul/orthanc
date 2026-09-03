package config

import "os"

const LOG_FILE = "/var/log/observer.log"

// UseEBPF selects the eBPF-based ProcessViewer instead of the polling one.
func UseEBPF() bool {
	return os.Getenv("OBSERVER_EBPF") == "1"
}

func GetEventCollectorHost() string {
	return os.Getenv("EVENT_COLLECTOR_HOST")
}
