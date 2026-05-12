// Package transcript provides functionality for formatting and writing transcript files.
package transcript

import (
	"fmt"
	"strings"
	"time"

	"github.com/theantichris/granola/internal/api"
)

// FormatTranscript formats transcript segments into a readable text format for file export.
func FormatTranscript(doc api.Document, segments []api.TranscriptSegment) string {
	if len(segments) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n")

	if doc.Title != "" {
		builder.WriteString(doc.Title)
		builder.WriteString("\n")
	}

	builder.WriteString("ID: ")
	builder.WriteString(doc.ID)
	builder.WriteString("\n")

	if doc.CreatedAt != "" {
		builder.WriteString("Created: ")
		builder.WriteString(doc.CreatedAt)
		builder.WriteString("\n")
	}

	if doc.UpdatedAt != "" {
		builder.WriteString("Updated: ")
		builder.WriteString(doc.UpdatedAt)
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("Segments: %d\n", len(segments)))
	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n\n")

	for _, segment := range segments {
		ts := parseTimestamp(segment.StartTime)
		label := speakerLabel(segment.Speaker)
		if ts != "" {
			builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", ts, label, segment.Text))
		} else {
			builder.WriteString(fmt.Sprintf("%s: %s\n", label, segment.Text))
		}
	}

	return builder.String()
}

// FormatTranscriptMarkdown formats transcript segments as Markdown lines for inline inclusion.
func FormatTranscriptMarkdown(segments []api.TranscriptSegment) string {
	if len(segments) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, segment := range segments {
		ts := parseTimestamp(segment.StartTime)
		label := speakerLabel(segment.Speaker)
		if ts != "" {
			builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", ts, label, segment.Text))
		} else {
			builder.WriteString(fmt.Sprintf("%s: %s\n", label, segment.Text))
		}
	}
	return builder.String()
}

// speakerLabel returns a human-readable label for a transcript speaker.
// On macOS: "microphone" = "You", "speaker" = "Speaker".
// On iOS: diarization_label ("Speaker A", "Speaker B", …) is used when present.
func speakerLabel(s api.TranscriptSpeaker) string {
	if s.DiarizationLabel != "" {
		return s.DiarizationLabel
	}
	if s.Source == "microphone" {
		return "You"
	}
	return "Speaker"
}

// parseTimestamp converts an ISO 8601 timestamp to HH:MM:SS format.
// Returns an empty string if the timestamp is empty or cannot be parsed.
func parseTimestamp(timestamp string) string {
	if timestamp == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return ""
	}
	return t.Format("15:04:05")
}
