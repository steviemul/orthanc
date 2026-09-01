package plugin

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/steviemul/orthanc-observer/event"
)

type BusyboxPlugin struct{}

func (bp BusyboxPlugin) CanHandle(e event.Event) bool {
	return strings.HasSuffix(e.Path, "/busybox")
}

func (bp BusyboxPlugin) ProcessEvent(e event.Event) *event.Evidence {

	args := readCmdLineArgs(e.PID)

	if len(args) == 0 {
		return nil
	}

	command := args[0]

	var commandArgs []string

	if len(args) > 0 {
		commandArgs = args[1:]
	}

	facts := map[string]string{
		"command": command,
		"args":    strings.Join(commandArgs, " "),
	}

	ev := &event.Evidence{
		Facts: facts,
	}

	return ev
}

func readCmdLineArgs(pid int) []string {

	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))

	if err != nil {
		fmt.Printf("unable to read command line args for pid %d", pid)
		return nil
	}

	parts := bytes.Split(data, []byte{0})
	args := make([]string, 0, len(parts))

	for _, part := range parts {
		if len(part) > 0 {
			args = append(args, string(part))
		}
	}

	return args
}
