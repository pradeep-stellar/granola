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

	builder.WriteString("Segments: ")
	builder.WriteString(fmt.Sprintf("%d", len(segments)))
	builder.WriteString("\n")

	builder.WriteString(strings.Repeat("=", 80))
	builder.WriteString("\n\n")

	for _, segment := range segments {
		startTime := parseTimestamp(segment.StartTimestamp)
		speaker := "System"
		if segment.Source == "microphone" {
			speaker = "You"
		}
		builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", startTime, speaker, segment.Text))
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
		startTime := parseTimestamp(segment.StartTimestamp)
		speaker := "System"
		if segment.Source == "microphone" {
			speaker = "You"
		}
		builder.WriteString(fmt.Sprintf("[%s] %s: %s\n", startTime, speaker, segment.Text))
	}
	return builder.String()
}

// parseTimestamp converts ISO 8601 timestamp to HH:MM:SS format.
func parseTimestamp(timestamp string) string {
	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return timestamp
	}
	return t.Format("15:04:05")
}
