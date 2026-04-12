package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnpackUnityI18NNative(t *testing.T) {
	inputPath := filepath.Join("unpack", "umbu", "unity-bundle-unwrap", "fr.bin")
	if _, err := os.Stat(inputPath); err != nil {
		t.Fatalf("missing test fixture %s: %v", inputPath, err)
	}

	outputPath := filepath.Join(t.TempDir(), "fr.json")
	if err := unpackUnityI18nNative(inputPath, outputPath); err != nil {
		t.Fatalf("unpackUnityI18nNative returned error: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed reading output file: %v", err)
	}

	var got unityI18NOutput
	if err := json.Unmarshal(content, &got); err != nil {
		t.Fatalf("failed parsing output json: %v", err)
	}

	if len(got.Entries) == 0 {
		t.Fatal("expected non-empty entries")
	}

	firstEntry, ok := got.Entries["1"]
	if !ok {
		t.Fatal("missing known key 1")
	}
	if !strings.Contains(firstEntry, "Commerce") {
		t.Fatalf("unexpected value for key 1: %q", firstEntry)
	}
}
