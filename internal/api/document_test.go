package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// ignoreRawFields is a cmp option that skips the RawFields map (raw JSON bytes for debug).
var ignoreRawFields = cmp.FilterPath(func(p cmp.Path) bool {
	return p.Last().String() == ".RawFields"
}, cmp.Ignore())

type errorTransport struct{}

func (e *errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("forced transport error")
}

func TestGetNotesWithAPIKey(t *testing.T) {
	t.Run("fetches metadata using cursor pagination", func(t *testing.T) {
		t.Parallel()

		page := 0
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth := r.Header.Get("Authorization"); auth != "Bearer grn_testkey" {
				t.Errorf("expected Authorization 'Bearer grn_testkey', got %q", auth)
			}
			w.WriteHeader(http.StatusOK)
			if page == 0 {
				page++
				_, _ = w.Write([]byte(`{"notes":[{"id":"not_abc","title":"First Meeting","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}],"hasMore":true,"cursor":"page2cursor"}`))
			} else {
				_, _ = w.Write([]byte(`{"notes":[{"id":"not_def","title":"Second Meeting","created_at":"2024-02-01T00:00:00Z","updated_at":"2024-02-02T00:00:00Z"}],"hasMore":false}`))
			}
		}))
		defer testServer.Close()

		httpClient := &http.Client{Transport: testServer.Client().Transport}

		actual, err := GetNotesWithAPIKey(testServer.URL, "grn_testkey", "", httpClient)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// List endpoint returns metadata only — Content is empty until GetNoteDetail is called.
		expected := []Document{
			{ID: "not_abc", Title: "First Meeting", CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z"},
			{ID: "not_def", Title: "Second Meeting", CreatedAt: "2024-02-01T00:00:00Z", UpdatedAt: "2024-02-02T00:00:00Z"},
		}

		if !cmp.Equal(actual, expected, ignoreRawFields) {
			t.Errorf("expected %v, got %v", expected, actual)
		}
	})

	t.Run("returns error for non-2xx status", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer testServer.Close()

		_, err := GetNotesWithAPIKey(testServer.URL, "grn_badkey", "", &http.Client{Transport: testServer.Client().Transport})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrDocumentAPI) {
			t.Errorf("expected %v, got %v", ErrDocumentAPI, err)
		}
	})

	t.Run("returns error for HTTP failure", func(t *testing.T) {
		t.Parallel()

		_, err := GetNotesWithAPIKey("http://test.dev", "grn_testkey", "", &http.Client{Transport: &errorTransport{}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrDocumentAPI) {
			t.Errorf("expected %v, got %v", ErrDocumentAPI, err)
		}
	})
}

func TestGetNoteDetail(t *testing.T) {
	t.Run("returns content from summary_markdown", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"not_abc","title":"Team Sync","summary_markdown":"## Notes\n\nAction items","summary_text":"Action items","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}`))
		}))
		defer testServer.Close()

		doc, err := GetNoteDetail(testServer.URL, "not_abc", "grn_testkey", false, testServer.Client())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if doc.Content != "## Notes\n\nAction items" {
			t.Errorf("expected summary_markdown as content, got %q", doc.Content)
		}
	})

	t.Run("falls back to summary_text when summary_markdown is empty", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"not_abc","title":"Team Sync","summary_markdown":"","summary_text":"Plain text summary","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z"}`))
		}))
		defer testServer.Close()

		doc, err := GetNoteDetail(testServer.URL, "not_abc", "grn_testkey", false, testServer.Client())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if doc.Content != "Plain text summary" {
			t.Errorf("expected summary_text as fallback, got %q", doc.Content)
		}
	})

	t.Run("includes transcript when withTranscript is true", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.RawQuery, "include=transcript") {
				t.Error("expected ?include=transcript in request URL")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"not_abc","title":"Team Sync","summary_markdown":"Notes","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z","transcript":[{"speaker":{"source":"microphone"},"text":"Hello","start_time":"2024-01-01T10:00:00Z","end_time":"2024-01-01T10:00:05Z"}]}`))
		}))
		defer testServer.Close()

		doc, err := GetNoteDetail(testServer.URL, "not_abc", "grn_testkey", true, testServer.Client())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(doc.Transcript) != 1 {
			t.Fatalf("expected 1 transcript segment, got %d", len(doc.Transcript))
		}
		if doc.Transcript[0].Text != "Hello" {
			t.Errorf("expected segment text 'Hello', got %q", doc.Transcript[0].Text)
		}
		if doc.Transcript[0].Speaker.Source != "microphone" {
			t.Errorf("expected speaker source 'microphone', got %q", doc.Transcript[0].Speaker.Source)
		}
	})

	t.Run("parses attendees into document", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"not_abc","title":"Team Sync","summary_markdown":"Notes","created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-02T00:00:00Z","attendees":[{"name":"Alice Smith","email":"alice@example.com"},{"name":"Bob Jones","email":"bob@example.com"}]}`))
		}))
		defer testServer.Close()

		doc, err := GetNoteDetail(testServer.URL, "not_abc", "grn_testkey", false, testServer.Client())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(doc.Attendees) != 2 {
			t.Fatalf("expected 2 attendees, got %d", len(doc.Attendees))
		}
		if doc.Attendees[0] != "Alice Smith" {
			t.Errorf("expected first attendee 'Alice Smith', got %q", doc.Attendees[0])
		}
		if doc.Attendees[1] != "Bob Jones" {
			t.Errorf("expected second attendee 'Bob Jones', got %q", doc.Attendees[1])
		}
	})

	t.Run("returns error for 404", func(t *testing.T) {
		t.Parallel()

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer testServer.Close()

		_, err := GetNoteDetail(testServer.URL, "not_missing", "grn_testkey", false, testServer.Client())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrDocumentAPI) {
			t.Errorf("expected %v, got %v", ErrDocumentAPI, err)
		}
	})
}
