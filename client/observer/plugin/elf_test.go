package plugin

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func withStubbedExtractor(t *testing.T, stub func(string) ([]string, error)) {
	t.Helper()

	original := extractSharedLibraries
	extractSharedLibraries = stub
	t.Cleanup(func() { extractSharedLibraries = original })
}

func TestReadSharedFilesWaitsForLinkerToSettle(t *testing.T) {
	// Simulate the dynamic linker progressively mapping more libraries
	// across successive reads of /proc/PID/maps, settling on the third read.
	snapshots := [][]string{
		{"/usr/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1"},
		{
			"/usr/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1",
			"/usr/lib/aarch64-linux-gnu/libc.so.6",
		},
		{
			"/usr/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1",
			"/usr/lib/aarch64-linux-gnu/libc.so.6",
		},
	}

	call := 0

	withStubbedExtractor(t, func(path string) ([]string, error) {
		idx := call
		if idx >= len(snapshots) {
			idx = len(snapshots) - 1
		}
		call++
		return snapshots[idx], nil
	})

	got := readSharedFiles(12)

	want := snapshots[len(snapshots)-1]

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	if call != 3 {
		t.Fatalf("expected extractor to be called 3 times (until settled), got %d", call)
	}
}

func TestReadSharedFilesGivesUpAfterMaxAttempts(t *testing.T) {
	// A snapshot that keeps changing every single read (e.g. libraries
	// loading slower than our retry budget) never settles, so we should
	// stop after mapsReadAttempts and return the last read rather than
	// blocking forever.
	call := 0

	withStubbedExtractor(t, func(path string) ([]string, error) {
		call++
		return []string{fmt.Sprintf("/lib/libcall%d.so", call)}, nil
	})

	got := readSharedFiles(12)

	if call != mapsReadAttempts {
		t.Fatalf("expected extractor to be called %d times, got %d", mapsReadAttempts, call)
	}

	want := []string{fmt.Sprintf("/lib/libcall%d.so", mapsReadAttempts)}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReadSharedFilesReturnsNilOnError(t *testing.T) {
	withStubbedExtractor(t, func(path string) ([]string, error) {
		return nil, errors.New("boom")
	})

	got := readSharedFiles(12)

	if got != nil {
		t.Fatalf("expected nil on error, got %v", got)
	}
}
