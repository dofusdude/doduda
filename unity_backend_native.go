package main

type nativeUnityUnpackBackend struct{}

func (nativeUnityUnpackBackend) Name() string { return UnityBackendNative }

func (nativeUnityUnpackBackend) Prepare(muteSpinner bool, headless bool) error {
	return nil
}

func (nativeUnityUnpackBackend) UnpackBundle(inputPath string, outputPath string) error {
	return unpackUnityBundleNative(inputPath, outputPath)
}

func (nativeUnityUnpackBackend) UnpackImages(inputDir string, outputDir string) error {
	return unpackUnityImagesNative(inputDir, outputDir)
}
