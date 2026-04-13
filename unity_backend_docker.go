package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/dofusdude/doduda/ui"
)

type dockerUnityUnpackBackend struct{}

func (dockerUnityUnpackBackend) Name() string { return UnityBackendDocker }

func (dockerUnityUnpackBackend) Prepare(muteSpinner bool, headless bool) error {
	return PullImages([]string{"stelzo/doduda-umbu:" + ARCH, "stelzo/assetstudio-cli:" + ARCH}, muteSpinner, headless)
}

func (dockerUnityUnpackBackend) UnpackBundle(inputPath string, outputPath string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	inputRawFileName := filepath.Base(inputPath)
	outputRawFileName := filepath.Base(outputPath)
	inputDir := filepath.Dir(inputPath)

	cmd := []string{
		"dotnet",
		"out/unity-bundle-unwrap.dll",
		path.Join("/app", "data", inputRawFileName),
		path.Join("/app", "data", outputRawFileName),
	}

	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "stelzo/doduda-umbu:" + ARCH,
		Cmd:   cmd,
		Volumes: map[string]struct{}{
			"/app/data": {},
		},
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/app/data", inputDir),
		},
		AutoRemove: true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}

func (dockerUnityUnpackBackend) UnpackImages(inputDir string, outputDir string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	imageName := "stelzo/assetstudio-cli:" + ARCH
	ctx := context.Background()
	bundles, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, bundle := range bundles {
		if bundle.IsDir() || !strings.HasSuffix(bundle.Name(), ".imagebundle") {
			continue
		}

		absInputPath := filepath.Join(inputDir, bundle.Name())
		cmd := []string{"./data", "--unity-version", "6000.0.41.58439"}
		uid := os.Getuid()
		gid := os.Getgid()

		containerConfig := &container.Config{
			Image: imageName,
			Cmd:   cmd,
			Volumes: map[string]struct{}{
				"/app/AssetStudio/data":     {},
				"/app/AssetStudio/ASExport": {},
			},
		}
		if uid >= 0 && gid >= 0 {
			containerConfig.User = strconv.Itoa(uid) + ":" + strconv.Itoa(gid)
		}

		resp, err := cli.ContainerCreate(ctx, containerConfig, &container.HostConfig{
			Binds: []string{
				fmt.Sprintf("%s:/app/AssetStudio/data", absInputPath),
				fmt.Sprintf("%s:/app/AssetStudio/ASExport", outputDir),
			},
			AutoRemove: true,
		}, nil, nil, "")
		if err != nil {
			return err
		}

		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return err
		}

		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-statusCh:
		}
	}

	return nil
}

func (dockerUnityUnpackBackend) UnpackI18n(inputPath string, outputPath string) error {
	return unpackUnityI18nNative(inputPath, outputPath)
}

func PullImages(images []string, muteSpinner bool, headless bool) error {
	feedbacks := make(chan string)

	var feedbackWg sync.WaitGroup
	if !muteSpinner {
		feedbackWg.Add(1)
		go func() {
			defer feedbackWg.Done()
			ui.Spinner("Docker", feedbacks, false, headless)
		}()
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, imageRef := range images {
		ctx := context.Background()
		if !muteSpinner {
			feedbacks <- "⬇️ " + imageRef
		}
		imageHandle, err := cli.ImagePull(ctx, imageRef, image.PullOptions{})
		if err != nil {
			return err
		}

		_, err = io.ReadAll(imageHandle)
		closeErr := imageHandle.Close()
		if err != nil {
			return fmt.Errorf("error finishing image handler: %w", err)
		}
		if closeErr != nil {
			return closeErr
		}
	}

	close(feedbacks)
	feedbackWg.Wait()

	return nil
}
