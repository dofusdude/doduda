package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	UnityBackendDocker = "docker"
	UnityBackendNative = "native"
)

type UnityUnpackBackend interface {
	Name() string
	Prepare(muteSpinner bool, headless bool) error
	UnpackBundle(inputPath string, outputPath string) error
	UnpackImages(inputDir string, outputDir string) error
}

func CurrentUnityUnpackBackend() (UnityUnpackBackend, error) {
	backend := selectedUnityBackend()
	switch backend {
	case UnityBackendDocker:
		return dockerUnityUnpackBackend{}, nil
	case UnityBackendNative:
		return nativeUnityUnpackBackend{}, nil
	default:
		return nil, fmt.Errorf("unknown unity backend %q", backend)
	}
}

func selectedUnityBackend() string {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("DODUDA_UNITY_BACKEND")))
	if backend == "" {
		return UnityBackendNative
	}

	switch backend {
	case UnityBackendDocker, UnityBackendNative:
		return backend
	default:
		return backend
	}
}
