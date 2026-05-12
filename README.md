# Granola CLI

```text
                                       _
                             _        | |
  __ _ _ __ __ _ _ __   ___ | | __ _  | |
 / _` | '__/ _` | '_ \ / _ \| |/ _` | | |
| (_| | | | (_| | | | | (_) | | (_| | | |
 \__, |_|  \__,_|_| |_|\___/|_|\__,_| | |
 |___/                                |_|
```

[![Go Version](https://img.shields.io/github/go-mod/go-version/theantichris/granola)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/theantichris/granola.svg)](https://pkg.go.dev/github.com/theantichris/granola)
[![Go Report Card](https://goreportcard.com/badge/github.com/theantichris/granola)](https://goreportcard.com/report/github.com/theantichris/granola)
[![Go](https://github.com/theantichris/granola/actions/workflows/go.yml/badge.svg)](https://github.com/theantichris/granola/actions/workflows/go.yml)
[![Markdown Lint](https://github.com/theantichris/granola/actions/workflows/markdown.yml/badge.svg)](https://github.com/theantichris/granola/actions/workflows/markdown.yml)
[![License](https://img.shields.io/github/license/theantichris/granola)](LICENSE)
[![Release](https://img.shields.io/github/v/release/theantichris/granola)](https://github.com/theantichris/granola/releases)

Export your [Granola](https://granola.ai) notes and transcripts to local files for backup, migration, or offline access.

## Why Use This?

- **Own Your Data** - Keep local copies of all your meeting notes
- **Full Transcripts** - Export complete, timestamped transcripts of your meetings
- **Backup & Migration** - Safeguard your notes or move them to other tools
- **Smart Updates** - Only exports new or changed content
- **Fast & Simple** - One command to export everything

## Installation

### Download Pre-Built Binary (Recommended)

1. Go to the [releases page](https://github.com/theantichris/granola/releases/latest)
2. Download the appropriate file for your operating system:
   - **macOS**: `granola_Darwin_x86_64.tar.gz` (Intel) or `granola_Darwin_arm64.tar.gz` (Apple Silicon)
   - **Linux**: `granola_Linux_x86_64.tar.gz`
   - **Windows**: `granola_Windows_x86_64.zip`
3. Extract the archive and move `granola` to a location in your PATH

### Install with Go

If you have Go installed:

```bash
go install github.com/theantichris/granola@latest
```

## Authentication

Both commands authenticate using a Granola API key. Set it as an environment variable:

```bash
export GRANOLA_API_KEY=your_api_key_here
```

Or add it to a `.env` file in your working directory:

```text
GRANOLA_API_KEY=your_api_key_here
```

## Quick Start

### Export Your Notes

Your notes are the AI-generated summaries and formatted content from Granola.

```bash
granola notes
```

Notes will be exported to a `notes/` directory as Markdown files.

To include the full meeting transcript inside each note under a `## Transcript` heading:

```bash
granola notes --transcript
```

### Export Your Transcripts

Transcripts are the raw, timestamped recordings of everything said in your meetings.

**Note:** Only meetings where audio recording was enabled will have transcript content.

```bash
granola transcripts
```

Transcripts will be exported to a `transcripts/` directory as plain text files.

## Common Options

### Custom Output Directory

```bash
# Export notes to a specific location
granola notes --output ~/Documents/MyNotes

# Export transcripts to a specific location
granola transcripts --output ~/Documents/MyTranscripts
```

### Set Default Configuration

Create a `.granola.toml` file in your home directory to avoid specifying flags every time:

```toml
# Notes configuration
output = "/Users/yourname/Documents/Notes"

# Transcripts configuration
transcript-output = "/Users/yourname/Documents/Transcripts"
```

Then simply run:

```bash
granola notes
granola transcripts
```

### Enable Debug Logging

```bash
granola notes --debug
granola transcripts --debug
```

## Command Reference

### `granola notes`

Fetches AI-generated notes from the Granola API and writes them as Markdown files.

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `./notes` | Output directory for Markdown files |
| `--transcript` | `false` | Append full transcript under `## Transcript` in each note |
| `--timeout` | `2m` | HTTP timeout for API requests |

### `granola transcripts`

Fetches transcripts from the Granola API and writes them as plain text files.

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `./transcripts` | Output directory for text files |
| `--timeout` | `2m` | HTTP timeout for API requests |

### Global Flags

| Flag | Description |
|------|-------------|
| `--debug` | Enable debug logging |
| `--config` | Path to config file (default: `$HOME/.granola.toml`) |

## What Gets Exported

### Notes (Markdown Files)

Each note becomes a separate `.md` file with:

- **YAML frontmatter** - ID and timestamps
- **Title** - As a top-level heading
- **Content** - AI-generated summary as Markdown

Example:

```markdown
---
id: abc-123
created: "2024-01-01T00:00:00Z"
updated: "2024-01-02T00:00:00Z"
---

# Team Sync Meeting

## Key Points

- First important point
- Second important point
```

With `--transcript`, a `## Transcript` section is appended:

```markdown
---
id: abc-123
created: "2024-01-01T00:00:00Z"
updated: "2024-01-02T00:00:00Z"
---

# Team Sync Meeting

## Key Points

- First important point

## Transcript

[14:00:04] System: Good morning everyone, how's it going?
[14:00:06] You: Good morning! Ready to start.
```

### Transcripts (Text Files)

Each transcript becomes a `.txt` file with:

- **Header** - Title, ID, timestamps, segment count
- **Timestamped dialogue** - `[HH:MM:SS] Speaker: Text`
- **Speaker labels** - `System` (other participants) or `You` (your microphone)

Example:

```text
================================================================================
Team Sync Meeting
ID: abc-123
Created: 2024-01-01T14:00:00Z
Updated: 2024-01-01T14:30:00Z
Segments: 142
================================================================================

[14:00:04] System: Good morning everyone, how's it going?
[14:00:06] You: Good morning! Ready to start.
```

## Troubleshooting

### "No API credentials configured"

- Set the `GRANOLA_API_KEY` environment variable or add it to a `.env` file
- Make sure the key is valid and has not expired

### "No transcripts found"

- Transcripts are only available for meetings where audio recording was enabled
- The API returns an empty transcript list for meetings without recordings

### "Permission denied"

- Make sure the output directory is writable
- Running as `sudo` is not needed

### Need More Help?

- Check the `--help` output: `granola notes --help`
- Enable debug logging with `--debug` for detailed request information
- [Open an issue](https://github.com/theantichris/granola/issues) on GitHub

---

## For Contributors & Developers

The sections below are for those who want to contribute to the project or build from source.

### Building from Source

**Requirements:**

- Go 1.23.1 or higher
- Git

**Clone and build:**

```bash
git clone https://github.com/theantichris/granola.git
cd granola
go build -o granola
```

**Cross-platform builds:**

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o granola-linux

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o granola-darwin

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o granola-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o granola.exe
```

### Project Structure

```text
granola/
├── cmd/                # Command implementations
│   ├── root.go         # Root command and configuration
│   ├── notes.go        # Notes export command
│   └── transcripts.go  # Transcripts export command
├── internal/           # Internal packages
│   ├── api/            # Granola API client and data models
│   ├── cache/          # Local cache file reader
│   ├── converter/      # Document to Markdown converter
│   ├── transcript/     # Transcript formatter
│   └── writer/         # File system operations
├── main.go             # Entry point
├── README.md           # This file
├── CLAUDE.md           # AI assistant guidelines
└── LICENSE             # MIT License
```

### Development Commands

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Run markdown linter (requires markdownlint-cli2)
markdownlint-cli2 "**/*.md" "#notes" "#transcripts"

# Run the CLI without building
go run main.go notes --help
```

### Testing

The project uses Go's standard testing framework with:

- **Unit tests** for individual components
- **Table-driven tests** for comprehensive coverage
- **Afero** for filesystem abstraction in tests
- **Parallel test execution** where possible

Run tests:

```bash
go test ./...           # All tests
go test -v ./...        # Verbose output
go test -cover ./...    # With coverage
```

### Releasing

Releases are automated using [GoReleaser](https://goreleaser.com/):

```bash
# Create and push a new tag
git tag v1.0.0
git push origin v1.0.0

# GitHub Actions will automatically build and publish the release
```

For local testing:

```bash
goreleaser release --snapshot --clean
```

### Contributing

Contributions are welcome! Here's how to help:

1. **Fork** the repository
2. **Create a branch** for your feature (`git checkout -b feature/amazing-feature`)
3. **Write tests** for your changes
4. **Ensure tests pass** (`go test ./...`)
5. **Commit your changes** (`git commit -m 'Add amazing feature'`)
6. **Push to your branch** (`git push origin feature/amazing-feature`)
7. **Open a Pull Request**

**Guidelines:**

- Follow existing code style
- Add tests for new functionality
- Update documentation as needed
- Keep PRs focused on a single change

### Key Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Charmbracelet Log](https://github.com/charmbracelet/log) - Structured logging
- [Afero](https://github.com/spf13/afero) - Filesystem abstraction

### Architecture

**Notes Export:**

1. Authenticate with `GRANOLA_API_KEY`
2. Fetch all notes from the Granola public API (cursor-based pagination)
3. Optionally fetch transcript for each note (`--transcript`)
4. Convert to Markdown with YAML frontmatter
5. Write files, skipping those that haven't changed

**Transcripts Export:**

1. Authenticate with `GRANOLA_API_KEY`
2. Fetch all notes from the Granola public API
3. Fetch transcript segments for each note
4. Format segments with timestamps and speaker labels
5. Write text files with metadata headers, skipping unchanged files

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Granola](https://granola.so) - The amazing note-taking app this tool exports from
- The Go team for the excellent language and tooling
- [Cobra](https://cobra.dev/) and [Viper](https://github.com/spf13/viper) for the CLI framework
- [Charmbracelet](https://charm.sh/) for beautiful terminal tools

## Support

For issues, questions, or feature requests:

- [Open an issue](https://github.com/theantichris/granola/issues) on GitHub
- Check existing issues for solutions
- Include debug output (`--debug`) when reporting problems

---

Built with love by the community
