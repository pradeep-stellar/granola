package cmd

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

func TestWriteNotes(t *testing.T) {
	t.Run("exports notes using GRANOLA_API_KEY", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if strings.Contains(r.URL.Path, "not_abc123") {
				// detail endpoint
				_, _ = w.Write([]byte(`{"id":"not_abc123","title":"Test Meeting","summary_markdown":"Meeting notes","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}`))
			} else {
				// list endpoint
				_, _ = w.Write([]byte(`{"notes":[{"id":"not_abc123","title":"Test Meeting","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}],"hasMore":false}`))
			}
		}))
		defer testServer.Close()

		viper.Reset()
		viper.Set("granola_api_key", "test_api_key_123")
		viper.Set("api_url", testServer.URL)
		viper.Set("timeout", time.Second)
		viper.Set("output", t.TempDir())

		logger := log.New(io.Discard)

		if err := writeNotes(logger); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when no credentials are configured", func(t *testing.T) {
		viper.Reset()

		err := writeNotes(log.New(io.Discard))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrNoCredentials) {
			t.Errorf("expected %v, got %v", ErrNoCredentials, err)
		}
	})
}
