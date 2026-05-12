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
			if err := viper.BindPFlag("timeout", cmd.Flags().Lookup("timeout")); err != nil {
				return fmt.Errorf("%w: %s", ErrNotesCmdInit, err)
			}
			if err := viper.BindPFlag("output", cmd.Flags().Lookup("output")); err != nil {
				return fmt.Errorf("%w: %s", ErrNotesCmdInit, err)
			}
			if err := viper.BindPFlag("transcript", cmd.Flags().Lookup("transcript")); err != nil {
				return fmt.Errorf("%w: %s", ErrNotesCmdInit, err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return writeNotes(logger)
		},
	}

	var timeout time.Duration
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "HTTP timeout for API requests, default 2 minutes")

	var output string
	cmd.Flags().StringVar(&output, "output", "./notes", "Output directory for exported Markdown files")

	var includeTranscript bool
	cmd.Flags().BoolVar(&includeTranscript, "transcript", false, "Include transcript under ## Transcript heading in each note")

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

	logger.Info("Authenticating with GRANOLA_API_KEY")
	fmt.Println("Using GRANOLA_API_KEY for authentication")
	fmt.Println("Fetching documents from Granola API...")
	logger.Info("Fetching documents from Granola API", "url", apiURL, "timeout", timeout)

	documents, err := api.GetNotesWithAPIKey(apiURL, apiKey, httpClient)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrDocumentExport, err)
	}

	logger.Info("Retrieved documents", "count", len(documents))

	if viper.GetBool("transcript") {
		fmt.Println("Fetching transcripts...")
		logger.Info("Fetching transcripts for all documents")
		for i, doc := range documents {
			segments, err := api.GetNoteTranscript(apiURL, doc.ID, apiKey, httpClient)
			if err != nil {
				logger.Warn("Failed to fetch transcript, skipping", "id", doc.ID, "error", err)
				continue
			}
			documents[i].Transcript = segments
		}
	}

	outputDir := viper.GetString("output")
	fmt.Printf("Exporting %d notes to %s...\n", len(documents), outputDir)
	logger.Info("Writing documents to Markdown files", "output", outputDir)

	if err := writer.Write(documents, outputDir, appFS); err != nil {
		return fmt.Errorf("%w: %s", ErrDocumentExport, err)
	}

	fmt.Println("✓ Export completed successfully")
	logger.Info("Export completed successfully", "files", len(documents))

	return nil
}
