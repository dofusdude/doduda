package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDownloadCDNRetry(t *testing.T) {
	for _, failure := range []string{"connection reset", "truncated body", "408", "429", "503", "404", "exhausted"} {
		t.Run(failure, func(t *testing.T) {
			calls, waits := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if calls > 1 && failure != "exhausted" {
					fmt.Fprint(w, "ok")
					return
				}
				switch failure {
				case "connection reset":
					conn, _, _ := w.(http.Hijacker).Hijack()
					conn.Close()
				case "truncated body":
					w.Header().Set("Content-Length", "100")
					fmt.Fprint(w, "short")
				case "408":
					w.WriteHeader(408)
				case "429":
					w.WriteHeader(429)
				case "404":
					w.WriteHeader(404)
				default:
					w.WriteHeader(503)
				}
			}))
			defer server.Close()
			body, err := downloadCDNWithRetry(server.Client(), server.URL, func(time.Duration) { waits++ })
			switch failure {
			case "404":
				if err == nil || calls != 1 || waits != 0 {
					t.Fatalf("permanent failure: calls=%d waits=%d err=%v", calls, waits, err)
				}
			case "exhausted":
				if err == nil || !strings.Contains(err.Error(), "after 4 attempts") || calls != 4 || waits != 3 {
					t.Fatalf("retry limit: calls=%d waits=%d err=%v", calls, waits, err)
				}
			default:
				if err != nil || string(body) != "ok" || calls != 2 || waits != 1 {
					t.Fatalf("recovery: calls=%d waits=%d body=%q err=%v", calls, waits, body, err)
				}
			}
		})
	}
}
