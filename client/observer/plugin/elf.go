package plugin

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/steviemul/orthanc-observer/event"
	"github.com/steviemul/orthanc-observer/io"
)

const (
	// mapsReadAttempts bounds how many times we re-read /proc/PID/maps
	// waiting for the dynamic linker to finish mapping shared libraries.
	mapsReadAttempts = 5
	mapsReadInterval = 20 * time.Millisecond
)

// extractSharedLibraries is a var (not a direct call) so tests can stub it
// to simulate a dynamic linker loading libraries across successive reads.
var extractSharedLibraries = io.ExtractSharedLibraries

type ElfPlugin struct{}

func (bp ElfPlugin) CanHandle(e event.Event) bool {
	return true
}

func (bp ElfPlugin) ProcessEvent(e event.Event) *event.Evidence {
	sharedFiles := readSharedFiles(e.PID)

	facts := map[string]string{
		"libraries": strings.Join(sharedFiles, ","),
	}

	return &event.Evidence{
		Facts: facts,
	}
}

// readSharedFiles reads /proc/PID/maps for the shared libraries a process
// has loaded. The exec tracepoint that triggers this fires as soon as the
// kernel loads the binary and its ELF interpreter (e.g. ld-linux), before
// that interpreter has actually resolved and mapped the binary's other
// shared libraries. Reading once here would race the dynamic linker and
// typically only capture the interpreter itself, so we re-read until two
// consecutive snapshots agree (linking has settled) or we run out of
// attempts.
func readSharedFiles(pid int) []string {

	mapsPath := fmt.Sprintf("/proc/%d/maps", pid)

	var sharedFiles []string

	for attempt := 0; attempt < mapsReadAttempts; attempt++ {
		current, err := extractSharedLibraries(mapsPath)

		if err != nil {
			fmt.Printf("unable to read mappings for pid %d", pid)
			return nil
		}

		if reflect.DeepEqual(current, sharedFiles) {
			return current
		}

		sharedFiles = current

		time.Sleep(mapsReadInterval)
	}

	return sharedFiles
}
