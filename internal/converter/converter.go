// Package converter provides functionality for converting Granola documents to Markdown format.
package converter

import (
	"fmt"
	"strings"

	"github.com/theantichris/granola/internal/api"
	"github.com/theantichris/granola/internal/transcript"
	"gopkg.in/yaml.v3"
)

// Metadata represents the YAML frontmatter for a Markdown file.
type Metadata struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	CreatedAt string   `yaml:"created"`
	UpdatedAt string   `yaml:"updated"`
	Folders   []string `yaml:"folders,omitempty"`
	Tags      []string `yaml:"tags,omitempty"`
}

// ToMarkdown converts a Document to Markdown format with YAML frontmatter.
func ToMarkdown(doc api.Document) (string, error) {
	metadata := Metadata{
		ID:        doc.ID,
		Title:     doc.Title,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
		Folders:   doc.Folders,
		Tags:      doc.Attendees,
	}

	yamlBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var builder strings.Builder

	builder.WriteString("---\n")
	builder.Write(yamlBytes)
	builder.WriteString("---\n\n")

	if doc.Title != "" {
		builder.WriteString("# ")
		builder.WriteString(doc.Title)
		builder.WriteString("\n\n")
	}

	if len(doc.Attendees) > 0 {
		for _, name := range doc.Attendees {
			builder.WriteString("[[")
			builder.WriteString(name)
			builder.WriteString("]] ")
		}
		builder.WriteString("\n\n")
	}

	content := strings.TrimSpace(doc.Content)
	if content != "" {
		builder.WriteString(content)
		builder.WriteString("\n")
	}

	if len(doc.Transcript) > 0 {
		builder.WriteString("\n## Transcript\n\n")
		builder.WriteString(transcript.FormatTranscriptMarkdown(doc.Transcript))
	}

	return builder.String(), nil
}
