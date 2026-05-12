package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type errorTransport struct{}

func (e *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("forced transport error")
}

func TestGetNotesWithAPIKey(t *testing.T) {
	t.Run("fetches notes using cursor pagination", func(t *testing.T) {
		t.Parallel()

		page := 0
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET request, got %s", r.Method)
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer grn_testkey" {
				t.Errorf("expected Authorization 'Bearer grn_testkey', got %q", auth)
			}
			w.WriteHeader(http.StatusOK)
			if page == 0 {
				page++
				_, _ = w.Write([]byte(`{"notes":[{"id":"not_abc","title":"First Meeting","summary_markdown":"## Notes\n\nSome content","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}],"hasMore":true,"cursor":"page2cursor"}`))
			} else {
				_, _ = w.Write([]byte(`{"notes":[{"id":"not_def","title":"Second Meeting","summary_markdown":"## Notes\n\nMore content","created_at":"2024-02-01T00:00:00Z","updated_at":"2024-02-02T00:00:00Z"}],"hasMore":false}`))
			}
		}))
		defer testServer.Close()

		httpClient := &http.Client{Transport: testServer.Client().Transport}

		actual, err := GetNotesWithAPIKey(testServer.URL, "grn_testkey", httpClient)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := []Document{
			{ID: "not_abc", Title: "First Meeting", Content: "## Notes\n\nSome content", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z"},
			{ID: "not_def", Title: "Second Meeting", Content: "## Notes\n\nMore content", CreatedAt: "2024-02-01T00:00:00Z", UpdatedAt: "2024-02-02T00:00:00Z"},
		}

		if !cmp.Equal(actual, expected) {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	})

	t.Run("returns error for non-2xx status", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer testServer.Close()

		httpClient := &http.Client{Transport: testServer.Client().Transport}

		_, err := GetNotesWithAPIKey(testServer.URL, "grn_badkey", httpClient)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrDocumentAPI) {
			t.Errorf("expected %v, got %v", ErrDocumentAPI, err)
		}
	})

	t.Run("returns error for HTTP failure", func(t *testing.T) {
		t.Parallel()

		httpClient := &http.Client{Transport: &errorTransport{}}

		_, err := GetNotesWithAPIKey("http://test.dev", "grn_testkey", httpClient)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrDocumentAPI) {
			t.Errorf("expected %v, got %v", ErrDocumentAPI, err)
		}
	})
}
