// Package writer provides functionality for writing Granola documents to Markdown files.
package writer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/theantichris/granola/internal/api"
	"github.com/theantichris/granola/internal/converter"
)

var invalidFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

// Write writes documents to Markdown files in the specified output directory.
// It only writes files if they don't exist or if the document's updated_at timestamp
// is newer than the existing file's modification time.
// When dryRun is true the files are not written; only what would be written is printed.
func Write(docs []api.Document, outputDir string, fs afero.Fs, logger *log.Logger, dryRun bool) error {
	if !dryRun {
		if err := fs.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	usedFilenames := make(map[string]int)
	written := 0
	skipped := 0

	for _, doc := range docs {
		filename := datePrefix(doc.CreatedAt) + sanitizeFilename(doc.Title, doc.ID)
		filename = makeUnique(filename, usedFilenames)
		usedFilenames[filename]++

		filePath := filepath.Join(outputDir, filename+".md")

		shouldWrite, err := shouldUpdateFile(fs, filePath, doc.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to check file status for %s: %w", filePath, err)
		}

		if !shouldWrite {
			skipped++
			continue
		}

		markdown, err := converter.ToMarkdown(doc)
		if err != nil {
			return fmt.Errorf("failed to convert document %s: %w", doc.ID, err)
		}

		preview := doc.Content
		if len(preview) > 30 {
			preview = preview[:30]
		}

		if dryRun {
			fmt.Printf("  [dry-run] %-50s %s\n", doc.Title, preview)
		} else {
			fmt.Printf("  %-50s %s\n", doc.Title, preview)
		}

		if logger.GetLevel() <= log.DebugLevel {
			logger.Debug("markdown preview", "file", filePath)
			fmt.Println(markdown)
		}

		if !dryRun {
			if err := afero.WriteFile(fs, filePath, []byte(markdown), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", filePath, err)
			}
		}

		written++
	}

	if dryRun {
		fmt.Printf("  [dry-run] %d would be written, %d up to date\n", written, skipped)
	}

	return nil
}

// shouldUpdateFile checks if a file should be written based on whether it exists
// and if the document's updated_at timestamp is newer than the file's modification time.
func shouldUpdateFile(fs afero.Fs, filePath string, updatedAt string) (bool, error) {
	// Check if file exists
	exists, err := afero.Exists(fs, filePath)
	if err != nil {
		return false, err
	}

	// If file doesn't exist, we should write it
	if !exists {
		return true, nil
	}

	// Parse the document's updated_at timestamp
	docUpdatedAt, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		// If we can't parse the timestamp, write the file to be safe
		return true, nil
	}

	// Get the file's modification time
	fileInfo, err := fs.Stat(filePath)
	if err != nil {
		return false, err
	}

	// Write the file if the document is newer than the existing file
	return docUpdatedAt.After(fileInfo.ModTime()), nil
}

// sanitizeFilename removes invalid characters from a filename and falls back to ID if empty.
func sanitizeFilename(title, id string) string {
	// Use title if available, otherwise use ID
	name := strings.TrimSpace(title)
	if name == "" {
		name = id
	}

	// Replace invalid characters with underscores
	name = invalidFileChars.ReplaceAllString(name, "_")

	// Replace multiple consecutive underscores with a single one
	name = regexp.MustCompile(`_+`).ReplaceAllString(name, "_")

	// Trim underscores from start and end
	name = strings.Trim(name, "_")

	// Ensure we have something
	if name == "" {
		name = "untitled"
	}

	// Limit length to 100 characters for filesystem compatibility
	if len(name) > 100 {
		name = name[:100]
	}

	return name
}

// datePrefix formats a RFC3339 timestamp as "YYYY-mm-dd-HHMM-" for use as a filename prefix.
// Returns an empty string if the timestamp cannot be parsed.
func datePrefix(createdAt string) string {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02-1504-")
}

// makeUnique appends a number to a filename if it already exists.
func makeUnique(filename string, used map[string]int) string {
	if count, exists := used[filename]; exists {
		return fmt.Sprintf("%s_%d", filename, count+1)
	}
	return filename
}