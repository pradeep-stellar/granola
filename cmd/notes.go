package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theantichris/granola/internal/api"
	"github.com/theantichris/granola/internal/writer"
)

var appFS = afero.NewOsFs()

const granolaPublicAPIURL = "https://public-api.granola.ai/v1/notes"

var (
	ErrNotesCmdInit   = errors.New("failed to initialize the notes command")
	ErrNoCredentials  = errors.New("no API credentials configured")
	ErrDocumentExport = errors.New("failed to export documents")
)

// NewNotesCmd creates a new notes command and binds its flags.
func NewNotesCmd(logger *log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "notes",
		Short:      "Export Granola notes to Markdown.",
		Long:       "Export Granola notes to Markdown files in the specified output directory.",
		SuggestFor: []string{"export"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, flag := range []string{"timeout", "output", "transcript", "since"} {
				if err := viper.BindPFlag(flag, cmd.Flags().Lookup(flag)); err != nil {
					return fmt.Errorf("%w: %s", ErrNotesCmdInit, err)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeNotes(logger)
		},
	}

	cmd.Flags().Duration("timeout", 2*time.Minute, "HTTP timeout for API requests")
	cmd.Flags().String("output", "./notes", "Output directory for exported Markdown files")
	cmd.Flags().Bool("transcript", false, "Include transcript under ## Transcript heading in each note")
	cmd.Flags().String("since", "", "Only fetch notes updated after this date (YYYY-MM-DD or ISO 8601)")

	return cmd
}

// writeNotes fetches notes from the Granola public API and writes them to Markdown files.
func writeNotes(logger *log.Logger) error {
	apiKey := strings.TrimSpace(viper.GetString("granola_api_key"))
	if apiKey == "" {
		return fmt.Errorf("%w: set GRANOLA_API_KEY environment variable", ErrNoCredentials)
	}

	apiURL := viper.GetString("api_url")
	if apiURL == "" {
		apiURL = granolaPublicAPIURL
	}

	timeout := viper.GetDuration("timeout")
	httpClient := &http.Client{Timeout: timeout}
	since := strings.TrimSpace(viper.GetString("since"))
	dryRun := viper.GetBool("dry-run")
	includeTranscript := viper.GetBool("transcript")

	if since != "" {
		fmt.Printf("Fetching notes updated after %s...\n", since)
	} else {
		fmt.Println("Fetching notes from Granola API...")
	}
	logger.Info("Fetching notes", "url", apiURL, "since", since, "timeout", timeout)

	documents, err := api.GetNotesWithAPIKey(apiURL, apiKey, since, httpClient)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrDocumentExport, err)
	}

	logger.Info("Retrieved notes", "count", len(documents))

	if len(documents) > 0 {
		first := documents[0]
		keys := make([]string, 0, len(first.RawFields))
		for k := range first.RawFields {
			keys = append(keys, k)
		}
		logger.Debug("first note API fields", "keys", keys)
	}

	total := len(documents)
	if includeTranscript {
		fmt.Printf("Fetching details and transcripts for %d notes...\n", total)
	} else {
		fmt.Printf("Fetching details for %d notes...\n", total)
	}

	for i := range documents {
		fmt.Printf("  [%d/%d] %s\n", i+1, total, documents[i].Title)
		detail, err := api.GetNoteDetail(apiURL, documents[i].ID, apiKey, includeTranscript, httpClient)
		if err != nil {
			logger.Warn("Failed to fetch note detail, skipping", "id", documents[i].ID, "error", err)
			continue
		}
		documents[i].Content = detail.Content
		documents[i].Folders = detail.Folders
		documents[i].Transcript = detail.Transcript
		documents[i].RawFields = detail.RawFields
	}

	outputDir := viper.GetString("output")
	if dryRun {
		fmt.Printf("Dry run — %d notes would be exported to %s:\n", total, outputDir)
	} else {
		fmt.Printf("Exporting %d notes to %s...\n", total, outputDir)
	}
	logger.Info("Writing notes", "output", outputDir, "dry_run", dryRun)

	if err := writer.Write(documents, outputDir, appFS, logger, dryRun); err != nil {
		return fmt.Errorf("%w: %s", ErrDocumentExport, err)
	}

	if !dryRun {
		fmt.Println("✓ Export completed successfully")
	}
	logger.Info("Done", "files", total)

	return nil
}
