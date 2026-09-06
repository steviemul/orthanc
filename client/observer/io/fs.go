package io

import (
	"fmt"
	"strings"

	"github.com/steviemul/orthanc-observer/event"
)

func GetFilePath(e event.Event, p string) string {

	if strings.HasPrefix("/", p) {
		return p
	}

	return e.Cwd + "/" + p
}

func GetEventFile(e event.Event, p string) string {

	filePath := GetFilePath(e, p)

	return fmt.Sprintf("/proc/%d/root%s", e.PID, filePath)
}
