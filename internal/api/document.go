package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	ErrDocumentAPI  = errors.New("failed to get documents")
	ErrDocumentJSON = errors.New("failed to unmarshal document JSON")
	ErrResponseBody = errors.New("failed to read response body")
	ErrHTTPRequest  = errors.New("failed to create HTTP request")
)

// Document contains a meeting note from Granola.
type Document struct {
	ID         string
	Title      string
	Content    string // summary_markdown from the public API
	CreatedAt  string
	UpdatedAt  string
	Transcript []TranscriptSegment // nil unless explicitly fetched
}

// TranscriptSegment represents a single speech segment in a note's transcript.
type TranscriptSegment struct {
	StartTimestamp string `json:"start_timestamp"`
	EndTimestamp   string `json:"end_timestamp"`
	Text           string `json:"text"`
	Source         string `json:"source"` // "system" or "microphone"
}

// TranscriptResponse represents the API response for a note's transcript.
type TranscriptResponse struct {
	Transcript []TranscriptSegment `json:"transcript"`
}

// PublicNoteOwner represents the owner of a note in the Granola public API.
type PublicNoteOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PublicNote represents a note returned by the Granola public API.
type PublicNote struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Owner           PublicNoteOwner `json:"owner"`
	Summary         string          `json:"summary"`
	SummaryMarkdown string          `json:"summary_markdown"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// PublicAPIListResponse represents the paginated response from the Granola public API.
type PublicAPIListResponse struct {
	Notes   []PublicNote `json:"notes"`
	HasMore bool         `json:"hasMore"`
	Cursor  string       `json:"cursor"`
}

// GetNotesWithAPIKey fetches all notes from the Granola public API using the provided API key.
// It uses cursor-based pagination to retrieve all results.
func GetNotesWithAPIKey(baseURL string, apiKey string, httpClient *http.Client) ([]Document, error) {
	var allDocuments []Document
	cursor := ""

	for {
		params := url.Values{"page_size": {"30"}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		reqURL := baseURL + "?" + params.Encode()

		httpRequest, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrHTTPRequest, err)
		}
		httpRequest.Header.Set("Authorization", "Bearer "+apiKey)

		response, err := httpClient.Do(httpRequest)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrDocumentAPI, err)
		}

		if response.StatusCode/100 != 2 {
			body, _ := io.ReadAll(response.Body)
			_ = response.Body.Close()
			preview := string(body)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return nil, fmt.Errorf("%w: status=%s, body=%s", ErrDocumentAPI, response.Status, preview)
		}

		responseBody, err := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrResponseBody, err)
		}

		var listResponse PublicAPIListResponse
		if err = json.Unmarshal(responseBody, &listResponse); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrDocumentJSON, err)
		}

		for _, note := range listResponse.Notes {
			allDocuments = append(allDocuments, Document{
				ID:        note.ID,
				Title:     note.Title,
				Content:   note.SummaryMarkdown,
				CreatedAt: note.CreatedAt,
				UpdatedAt: note.UpdatedAt,
			})
		}

		if !listResponse.HasMore {
			break
		}
		cursor = listResponse.Cursor
	}

	return allDocuments, nil
}

// GetNoteTranscript fetches the transcript segments for a specific note.
func GetNoteTranscript(baseURL string, noteID string, apiKey string, httpClient *http.Client) ([]TranscriptSegment, error) {
	reqURL := baseURL + "/" + noteID + "/transcript"

	httpRequest, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHTTPRequest, err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDocumentAPI, err)
	}

	if response.StatusCode/100 != 2 {
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return nil, fmt.Errorf("%w: status=%s, body=%s", ErrDocumentAPI, response.Status, preview)
	}

	responseBody, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrResponseBody, err)
	}

	var transcriptResponse TranscriptResponse
	if err = json.Unmarshal(responseBody, &transcriptResponse); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrDocumentJSON, err)
	}

	return transcriptResponse.Transcript, nil
}
