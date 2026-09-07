package plugin

import (
	"fmt"
	"strings"

	"github.com/steviemul/orthanc-observer/event"
	"github.com/steviemul/orthanc-observer/io"
)

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

func readSharedFiles(pid int) []string {

	sharedFiles, err := io.ExtractSharedLibraries(fmt.Sprintf("/proc/%d/maps", pid))

	if err != nil {
		fmt.Printf("unable to read mappings for pid %d", pid)
		return nil
	}

	return sharedFiles
}
