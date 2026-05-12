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
	Content    string // summary_markdown, falling back to summary_text
	CreatedAt  string
	UpdatedAt  string
	Folders    []string            // folder names from folder_membership
	Attendees  []string            // attendee names
	Transcript []TranscriptSegment // nil unless fetched with include=transcript
	RawFields  map[string]json.RawMessage
}

// TranscriptSpeaker identifies the speaker of a transcript segment.
type TranscriptSpeaker struct {
	// Source is "microphone" (user's mic) or "speaker" (other meeting audio).
	Source string `json:"source"`
	// DiarizationLabel is only present on iOS recordings, e.g. "Speaker A".
	DiarizationLabel string `json:"diarization_label,omitempty"`
}

// TranscriptSegment represents a single speech segment in a note's transcript.
type TranscriptSegment struct {
	Speaker   TranscriptSpeaker `json:"speaker"`
	Text      string            `json:"text"`
	StartTime string            `json:"start_time"`
	EndTime   string            `json:"end_time"`
}

// FolderMembership represents a folder that a note belongs to.
type FolderMembership struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id"`
}

// PublicNoteOwner represents the owner of a note.
type PublicNoteOwner struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Attendee represents a meeting participant.
type Attendee struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PublicNote represents a note as returned by the Granola public API.
// The list endpoint populates only the metadata fields; content, folders, and
// transcript require a detail call to GET /v1/notes/{id}.
type PublicNote struct {
	ID               string              `json:"id"`
	Object           string              `json:"object"`
	Title            string              `json:"title"`
	Owner            PublicNoteOwner     `json:"owner"`
	CreatedAt        string              `json:"created_at"`
	UpdatedAt        string              `json:"updated_at"`
	SummaryMarkdown  string              `json:"summary_markdown"`
	SummaryText      string              `json:"summary_text"`
	FolderMembership []FolderMembership  `json:"folder_membership"`
	Attendees        []Attendee          `json:"attendees"`
	Transcript       []TranscriptSegment `json:"transcript"`

	// RawFields captures every key in the response for debug inspection.
	RawFields map[string]json.RawMessage `json:"-"`
}

func (n *PublicNote) UnmarshalJSON(data []byte) error {
	type Alias PublicNote
	if err := json.Unmarshal(data, (*Alias)(n)); err != nil {
		return err
	}
	return json.Unmarshal(data, &n.RawFields)
}

// content returns summary_markdown when available, falling back to summary_text.
func (n PublicNote) content() string {
	if n.SummaryMarkdown != "" {
		return n.SummaryMarkdown
	}
	return n.SummaryText
}

// folderNames extracts just the folder names from folder_membership.
func (n PublicNote) folderNames() []string {
	if len(n.FolderMembership) == 0 {
		return nil
	}
	names := make([]string, len(n.FolderMembership))
	for i, f := range n.FolderMembership {
		names[i] = f.Name
	}
	return names
}

// attendeeNames extracts non-empty attendee names.
func (n PublicNote) attendeeNames() []string {
	var names []string
	for _, a := range n.Attendees {
		if a.Name != "" {
			names = append(names, a.Name)
		}
	}
	return names
}

// PublicAPIListResponse represents the paginated list response from the Granola API.
type PublicAPIListResponse struct {
	Notes   []PublicNote `json:"notes"`
	HasMore bool         `json:"hasMore"`
	Cursor  string       `json:"cursor"`
}

// GetNotesWithAPIKey fetches all notes using cursor-based pagination.
// updatedAfter is an optional ISO 8601 date/datetime; when non-empty only notes
// updated after that point are returned. The list endpoint returns metadata only
// (no content, folders, or transcript).
func GetNotesWithAPIKey(baseURL, apiKey, updatedAfter string, httpClient *http.Client) ([]Document, error) {
	var allDocuments []Document
	cursor := ""

	for {
		params := url.Values{"page_size": {"30"}}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		if updatedAfter != "" {
			params.Set("updated_after", updatedAfter)
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
				CreatedAt: note.CreatedAt,
				UpdatedAt: note.UpdatedAt,
				RawFields: note.RawFields,
			})
		}

		if !listResponse.HasMore {
			break
		}
		cursor = listResponse.Cursor
	}

	return allDocuments, nil
}

// GetNoteDetail fetches the full details of a single note from GET /v1/notes/{id}.
// Pass withTranscript=true to append ?include=transcript and retrieve transcript
// segments in the same call.
func GetNoteDetail(baseURL, noteID, apiKey string, withTranscript bool, httpClient *http.Client) (Document, error) {
	reqURL := baseURL + "/" + noteID
	if withTranscript {
		reqURL += "?include=transcript"
	}

	httpRequest, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return Document{}, fmt.Errorf("%w: %s", ErrHTTPRequest, err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return Document{}, fmt.Errorf("%w: %s", ErrDocumentAPI, err)
	}

	if response.StatusCode == http.StatusNotFound {
		_ = response.Body.Close()
		return Document{}, fmt.Errorf("%w: note %s not found", ErrDocumentAPI, noteID)
	}

	if response.StatusCode/100 != 2 {
		body, _ := io.ReadAll(response.Body)
		_ = response.Body.Close()
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		return Document{}, fmt.Errorf("%w: status=%s, body=%s", ErrDocumentAPI, response.Status, preview)
	}

	responseBody, err := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if err != nil {
		return Document{}, fmt.Errorf("%w: %s", ErrResponseBody, err)
	}

	var note PublicNote
	if err = json.Unmarshal(responseBody, &note); err != nil {
		return Document{}, fmt.Errorf("%w: %s", ErrDocumentJSON, err)
	}

	return Document{
		ID:         note.ID,
		Title:      note.Title,
		Content:    note.content(),
		CreatedAt:  note.CreatedAt,
		UpdatedAt:  note.UpdatedAt,
		Folders:    note.folderNames(),
		Attendees:  note.attendeeNames(),
		Transcript: note.Transcript,
		RawFields:  note.RawFields,
	}, nil
}

// GetNoteTranscript fetches the transcript segments for a note.
// It uses ?include=transcript on the detail endpoint to retrieve both in one call.
func GetNoteTranscript(baseURL, noteID, apiKey string, httpClient *http.Client) ([]TranscriptSegment, error) {
	doc, err := GetNoteDetail(baseURL, noteID, apiKey, true, httpClient)
	if err != nil {
		return nil, err
	}
	return doc.Transcript, nil
}
