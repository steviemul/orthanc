package io

import (
	"bufio"
	"fmt"
	"os"
	"sort"
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

func ExtractSharedLibraries(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Use a map as a set to automatically deduplicate library paths
	seenLibs := make(map[string]bool)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// strings.Fields automatically splits on any sequential whitespace
		fields := strings.Fields(line)

		// A line must have at least 6 fields to contain a file path.
		// Layout: [0]address [1]perms [2]offset [3]dev [4]inode [5...]path
		if len(fields) < 6 {
			continue
		}

		// Reconstruct the path. Because strings.Fields splits on all spaces,
		// if a path contains spaces, it will be split across fields[5:]
		path := strings.Join(fields[5:], " ")

		// Filter for lines containing shared libraries
		if strings.Contains(path, ".so") {
			seenLibs[path] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert the map keys into a slice
	var libraries []string
	for lib := range seenLibs {
		libraries = append(libraries, lib)
	}

	// Sort alphabetically for clean, predictable output
	sort.Strings(libraries)

	return libraries, nil
}
