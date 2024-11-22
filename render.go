package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/dofusdude/doduda/ui"
)

func Render(inputDir string, outputDir string, incrementalParts []string, resolution int, headless bool) error {
	updateChan := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ui.Spinner("Incremental", updateChan, false, headless)
		if !isChannelClosed(updateChan) {
			close(updateChan)
		}
	}()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	if isChannelClosed(updateChan) {
		os.Exit(1)
	}
	updateChan <- "Pulling image"

	ctx := context.Background()
	swfToSvg, err := cli.ImagePull(ctx, "stelzo/swf-to-svg", image.PullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	swfToSvg.Close()

	svgToPng, err := cli.ImagePull(ctx, "stelzo/svg-to-png", image.PullOptions{})
	if err != nil {
		log.Fatal(err)
	}
	svgToPng.Close()

	swfFiles, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatal(err)
	}

	if len(incrementalParts) != 0 {
		owner := incrementalParts[0]
		repo := incrementalParts[1]
		filename := incrementalParts[2] + ".tar.gz"

		if isChannelClosed(updateChan) {
			os.Exit(1)
		}
		updateChan <- "Checking latest release"

		releaseApiResponse, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo))
		if err != nil {
			return err
		}

		releaseApiResponseBody, err := io.ReadAll(releaseApiResponse.Body)
		if err != nil {
			return err
		}

		var v map[string]interface{}
		err = json.Unmarshal(releaseApiResponseBody, &v)
		if err != nil {
			return err
		}

		assets := v["assets"].([]interface{})
		found := false
		for _, asset := range assets {
			assetMap := asset.(map[string]interface{})
			if assetMap["name"].(string) == filename {
				found = true
				assetUrl := assetMap["browser_download_url"].(string)

				if isChannelClosed(updateChan) {
					os.Exit(1)
				}
				updateChan <- "loading latest " + filename

				imagesResponse, err := http.Get(assetUrl)
				if err != nil {
					log.Fatal(err)
				}

				err = ExtractTarGz("", imagesResponse.Body)
				if err != nil {
					return err
				}
			}
		}

		if !found {
			log.Fatal("Could not find the specified file in the latest release")
		}
	}

	if !isChannelClosed(updateChan) {
		close(updateChan)
	}

	wg.Wait()

	progressChan := make(chan bool, len(swfFiles))
	wg.Add(1)
	go func() {
		defer wg.Done()
		ui.Progress("Rendering", len(swfFiles), progressChan, 0, true, headless)
	}()

	for _, swfFile := range swfFiles {
		if swfFile.IsDir() || !strings.HasSuffix(swfFile.Name(), ".swf") {
			progressChan <- true
			continue
		}

		rawFileName := strings.TrimSuffix(swfFile.Name(), ".swf")
		svgFileName := fmt.Sprintf("%s.svg", rawFileName)
		resultFileName := fmt.Sprintf("%s-%d.png", rawFileName, resolution)

		absInputPath := filepath.Join(inputDir, swfFile.Name())
		tmpOutputPath := filepath.Join(inputDir, resultFileName)
		absOutputPath := filepath.Join(outputDir, resultFileName)

		if _, err := os.Stat(absOutputPath); err == nil {
			progressChan <- true
			continue // skip already rendered files
		}

		mountPath := filepath.Dir(absInputPath)

		cmd := []string{
			filepath.Join("data", swfFile.Name()),
			filepath.Join("data", svgFileName),
		}

		resp, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "stelzo/swf-to-svg",
			Cmd:   cmd,
			Volumes: map[string]struct{}{
				"/app/data": {},
			},
		}, &container.HostConfig{
			Binds:      []string{fmt.Sprintf("%s:/app/data", mountPath)},
			AutoRemove: true,
		}, nil, nil, "")
		if err != nil {
			log.Fatal(err)
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

		cmd = []string{
			filepath.Join("data", svgFileName),
			filepath.Join("data", resultFileName),
			strconv.Itoa(resolution),
		}

		resp, err = cli.ContainerCreate(ctx, &container.Config{
			Image: "stelzo/svg-to-png",
			Cmd:   cmd,
			Volumes: map[string]struct{}{
				"/app/data": {},
			},
		}, &container.HostConfig{
			Binds:      []string{fmt.Sprintf("%s:/app/data", mountPath)},
			AutoRemove: true,
		}, nil, nil, "")
		if err != nil {
			log.Fatal(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return err
		}

		statusCh, errCh = cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-statusCh:
		}

		err = os.Rename(tmpOutputPath, absOutputPath)
		if err != nil {
			log.Warn("File " + swfFile.Name() + " could not be converted")
		}
		progressChan <- true
	}

	wg.Wait()

	return nil
}

// from Armatorix https://codereview.stackexchange.com/questions/272457/decompress-tar-gz-file-in-go
func ExtractTarGz(baseDir string, gzipStream io.Reader) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filepath.Join(baseDir, header.Name), 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %w", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(filepath.Join(baseDir, header.Name))
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %w", err)
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %w", err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("ExtractTarGz: Close() failed: %w", err)
			}
		default:
			return fmt.Errorf("ExtractTarGz: uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}
	if err != io.EOF {
		return fmt.Errorf("ExtractTarGz: Next() failed: %w", err)
	}
	return nil
}
