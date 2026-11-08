package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var cdnClient = &http.Client{Timeout: 2 * time.Minute}

func downloadCDN(url string) ([]byte, error) {
	return downloadCDNWithRetry(cdnClient, url, time.Sleep)
}

// Retry transient CDN failures, including interrupted response bodies. Permanent
// HTTP errors are returned immediately; every response body is closed.
func downloadCDNWithRetry(client *http.Client, url string, sleep func(time.Duration)) ([]byte, error) {
	const attempts = 4
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			sleep(time.Second << (attempt - 1))
		}
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
			if resp.StatusCode != http.StatusRequestTimeout && resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
				return nil, lastErr
			}
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err == nil {
			return body, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("download %s failed after %d attempts: %w", url, attempts, lastErr)
}
