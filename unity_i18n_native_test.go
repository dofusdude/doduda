package main

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUnpackUnityI18NNative(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "fr.bin")
	if err := os.WriteFile(inputPath, testUnityI18NNativeFixture(), 0o600); err != nil {
		t.Fatalf("failed writing test fixture: %v", err)
	}

	outputPath := filepath.Join(tmpDir, "fr.json")
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
	if firstEntry != "Commerce" {
		t.Fatalf("unexpected value for key 1: %q", firstEntry)
	}
}

func testUnityI18NNativeFixture() []byte {
	data := make([]byte, 0, 24)
	buf := make([]byte, 4)

	// Version byte + 2 reserved bytes.
	data = append(data, 0, 0, 0)

	// Integer table count.
	binary.LittleEndian.PutUint32(buf, 1)
	data = append(data, buf...)

	// Entry key.
	binary.LittleEndian.PutUint32(buf, 1)
	data = append(data, buf...)

	// Entry string offset (version+reserved+count+entry table = 15).
	binary.LittleEndian.PutUint32(buf, 15)
	data = append(data, buf...)

	// String data.
	data = append(data, 8)
	data = append(data, []byte("Commerce")...)

	return data
}
