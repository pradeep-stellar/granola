package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/granola/internal/api"
	"github.com/theantichris/granola/internal/transcript"
)

var (
	ErrTranscriptCmdInit = errors.New("failed to initialize the transcripts command")
	ErrTranscriptExport  = errors.New("failed to export transcripts")
	invalidCharsRegex    = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
)

// NewTranscriptsCmd creates a new transcripts command and binds its flags.
func NewTranscriptsCmd(logger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transcripts",
		Short: "Export Granola transcripts to text files.",
		Long:  "Export raw Granola transcripts with timestamps to plain text files in the specified output directory.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, flag := range []string{"transcript-output", "transcript-timeout", "transcript-since"} {
				f := strings.TrimPrefix(flag, "transcript-")
				if err := viper.BindPFlag(flag, cmd.Flags().Lookup(f)); err != nil {
					return fmt.Errorf("%w: %s", ErrTranscriptCmdInit, err)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeTranscripts(logger)
		},
	}

	cmd.Flags().String("output", "./transcripts", "Output directory for exported transcript files")
	cmd.Flags().Duration("timeout", 2*time.Minute, "HTTP timeout for API requests")
	cmd.Flags().String("since", "", "Only fetch notes updated after this date (YYYY-MM-DD or ISO 8601)")

	return cmd
}

// writeTranscripts fetches notes from the API, retrieves transcripts for each, and writes them to text files.
func writeTranscripts(logger *log.Logger) error {
	apiKey := strings.TrimSpace(viper.GetString("granola_api_key"))
	if apiKey == "" {
		return fmt.Errorf("%w: set GRANOLA_API_KEY environment variable", ErrNoCredentials)
	}

	apiURL := viper.GetString("api_url")
	if apiURL == "" {
		apiURL = granolaPublicAPIURL
	}

	timeout := viper.GetDuration("transcript-timeout")
	httpClient := &http.Client{Timeout: timeout}
	since := strings.TrimSpace(viper.GetString("transcript-since"))
	dryRun := viper.GetBool("dry-run")

	if since != "" {
		fmt.Printf("Fetching notes updated after %s...\n", since)
	} else {
		fmt.Println("Fetching notes from Granola API...")
	}
	logger.Info("Fetching notes", "url", apiURL, "since", since)

	documents, err := api.GetNotesWithAPIKey(apiURL, apiKey, since, httpClient)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTranscriptExport, err)
	}

	logger.Info("Retrieved notes", "count", len(documents))

	outputDir := viper.GetString("transcript-output")
	total := len(documents)

	if dryRun {
		fmt.Printf("Dry run — fetching transcripts for %d notes (nothing will be written)...\n", total)
	} else {
		fmt.Printf("Fetching transcripts for %d notes, exporting to %s...\n", total, outputDir)
	}

	if !dryRun {
		if err := appFS.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("%w: failed to create output directory: %s", ErrTranscriptExport, err)
		}
	}

	usedFilenames := make(map[string]bool)
	count := 0

	for i, doc := range documents {
		title := doc.Title
		if title == "" {
			title = doc.ID
		}
		filename := transcriptDatePrefix(doc.CreatedAt) + sanitizeFilename(title)
		filename = makeUnique(filename, usedFilenames)
		usedFilenames[filename] = true
		filePath := filepath.Join(outputDir, filename+".txt")

		if !dryRun && !shouldUpdateFileByDoc(doc, filePath) {
			continue
		}

		fmt.Printf("  [%d/%d] %s\n", i+1, total, doc.Title)
		logger.Debug("Fetching transcript", "id", doc.ID, "title", doc.Title)

		segments, err := api.GetNoteTranscript(apiURL, doc.ID, apiKey, httpClient)
		if err != nil {
			logger.Warn("Failed to fetch transcript, skipping", "id", doc.ID, "error", err)
			continue
		}

		if len(segments) == 0 {
			continue
		}

		content := transcript.FormatTranscript(doc, segments)
		if content == "" {
			continue
		}

		if dryRun {
			fmt.Printf("    [dry-run] would write %s (%d segments)\n", filePath, len(segments))
			count++
			continue
		}

		if err := afero.WriteFile(appFS, filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("%w: failed to write file %s: %s", ErrTranscriptExport, filePath, err)
		}

		count++
	}

	if dryRun {
		fmt.Printf("  [dry-run] %d transcripts would be written\n", count)
	} else {
		fmt.Println("✓ Export completed successfully")
		logger.Info("Export completed successfully", "files", count)
	}

	return nil
}

// transcriptDatePrefix formats a RFC3339 timestamp as "YYYY-mm-dd-HHMM-" for use as a filename prefix.
func transcriptDatePrefix(createdAt string) string {
	t, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02-1504-")
}

// sanitizeFilename removes invalid characters and limits length.
func sanitizeFilename(name string) string {
	name = invalidCharsRegex.ReplaceAllString(name, "_")
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}

// makeUnique ensures filename is unique by appending a number if needed.
func makeUnique(filename string, used map[string]bool) string {
	if !used[filename] {
		return filename
	}
	counter := 2
	for {
		uniqueName := fmt.Sprintf("%s_%d", filename, counter)
		if !used[uniqueName] {
			return uniqueName
		}
		counter++
	}
}

// shouldUpdateFileByDoc checks if the file needs to be updated based on document timestamps.
func shouldUpdateFileByDoc(doc api.Document, filePath string) bool {
	fileInfo, err := appFS.Stat(filePath)
	if err != nil {
		return true
	}
	docUpdated, err := time.Parse(time.RFC3339, doc.UpdatedAt)
	if err != nil {
		return true
	}
	return docUpdated.After(fileInfo.ModTime())
}
