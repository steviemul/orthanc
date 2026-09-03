package plugin

import (
	"strings"

	"github.com/steviemul/orthanc-observer/event"
)

type JavaPlugin struct{}

func (bp JavaPlugin) CanHandle(e event.Event) bool {
	return e.Process == "java" || strings.HasSuffix(e.Path, "/java")
}

func (bp JavaPlugin) ProcessEvent(e event.Event) *event.Evidence {

	args := readCmdLineArgs(e.PID)

	if len(args) == 0 {
		return nil
	}

	facts := map[string]string{
		"args": strings.Join(args, " "),
	}

	jarPath := getJarPath(args)

	if jarPath != "" {
		facts["jarPath"] = jarPath
	}

	ev := &event.Evidence{
		Facts: facts,
	}

	return ev
}

func getJarPath(javaArgs []string) string {

	for i, arg := range javaArgs {

		if "-jar" == arg {
			return javaArgs[i+1]
		}
	}

	return ""
}
