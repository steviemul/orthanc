package io

import (
	"reflect"
	"testing"
)

func TestExtractSharedLibraries(t *testing.T) {
	libraries, err := ExtractSharedLibraries("testdata/maps_sample.txt")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"/opt/java/openjdk/lib/libjli.so",
		"/usr/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1",
		"/usr/lib/aarch64-linux-gnu/libc.so.6",
		"/usr/lib/aarch64-linux-gnu/libdl.so.2",
		"/usr/lib/aarch64-linux-gnu/libpthread.so.0",
	}

	if !reflect.DeepEqual(libraries, expected) {
		t.Fatalf("got %v, want %v", libraries, expected)
	}
}

func TestExtractSharedLibrariesMissingFile(t *testing.T) {
	_, err := ExtractSharedLibraries("testdata/does_not_exist.txt")

	if err == nil {
		t.Fatal("expected an error for a missing file, got nil")
	}
}
