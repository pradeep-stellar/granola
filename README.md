# granola-cli

```text
                                       _
                             _        | |
  __ _ _ __ __ _ _ __   ___ | | __ _  | |
 / _` | '__/ _` | '_ \ / _ \| |/ _` | | |
| (_| | | | (_| | | | | (_) | | (_| | | |
 \__, |_|  \__,_|_| |_|\___/|_|\__,_| | |
 |___/                                |_|
```

[![Go](https://github.com/pradeep-stellar/granola/actions/workflows/go.yml/badge.svg)](https://github.com/pradeep-stellar/granola/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/pradeep-stellar/granola)](https://github.com/pradeep-stellar/granola/releases)
[![License](https://img.shields.io/github/license/pradeep-stellar/granola)](LICENSE)

Your [Granola](https://granola.ai) notes, as plain files on your machine.

---

Granola does a great job capturing meetings. This tool gets those notes out — as Markdown files
you actually own, in a folder you choose, ready for Obsidian, Notion, git, or whatever comes next.

## What it does

```bash
granola notes        # AI-generated notes → Markdown files
granola transcripts  # Raw meeting transcripts → plain text files
```

Files are named `YYYY-MM-DD-HHMM-Title.md` so they sort naturally. Only new or updated
notes are written on subsequent runs, so re-running is fast.

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap pradeep-stellar/brew
brew install granola
```

### Download a binary

Grab the latest from the [releases page](https://github.com/pradeep-stellar/granola/releases/latest):

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `granola_Darwin_arm64.tar.gz` |
| macOS (Intel) | `granola_Darwin_x86_64.tar.gz` |
| Linux | `granola_Linux_x86_64.tar.gz` |
| Windows | `granola_Windows_x86_64.zip` |

### Go

```bash
go install github.com/pradeep-stellar/granola@latest
```

## Setup

Set your Granola API key:

```bash
export GRANOLA_API_KEY=your_api_key_here
```

Or drop it in a `.env` file in your working directory:

```text
GRANOLA_API_KEY=your_api_key_here
```

## Usage

### Export notes

```bash
granola notes
```

Exports AI-generated summaries to `./notes/` as `.md` files with YAML frontmatter:

```markdown
---
id: not_abc123
title: Team Sync
created: "2024-06-01T14:00:00Z"
updated: "2024-06-01T14:35:00Z"
folders:
  - Work
  - Q2 Planning
tags:
  - Alice Smith
  - Bob Jones
---

# Team Sync

[[Alice Smith]] [[Bob Jones]]

## Key decisions

- Launched the new onboarding flow next Tuesday
- Moved standup to 9:30am
```

Attendees appear as `[[wikilinks]]` right after the title (clickable in Obsidian and similar tools)
and as a `tags:` list in the frontmatter for easy filtering.

Want the full transcript inline too?

```bash
granola notes --transcript
```

That appends a `## Transcript` section to each note with timestamped dialogue.

### Export transcripts

```bash
granola transcripts
```

Exports raw transcripts to `./transcripts/` as `.txt` files — one file per meeting,
with a header block and `[HH:MM:SS] Speaker: text` lines.

> Only meetings where audio recording was enabled will have transcript content.

### Useful flags

```bash
# Only fetch notes newer than a date (great for incremental runs)
granola notes --since 2024-06-01

# Preview what would be written without touching the filesystem
granola notes --dry-run

# Write to a different folder
granola notes --output ~/Documents/Notes

# See what's happening under the hood
granola notes --debug
```

### Save defaults in a config file

Create `~/.granola.toml` to avoid repeating yourself:

```toml
output = "/Users/you/Documents/Notes"
transcript-output = "/Users/you/Documents/Transcripts"
```

## Flag reference

### `granola notes`

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `./notes` | Output directory |
| `--transcript` | `false` | Append full transcript to each note |
| `--since` | — | Only fetch notes updated after this date |
| `--dry-run` | `false` | Show what would be written, without writing |
| `--timeout` | `2m` | HTTP timeout |

### `granola transcripts`

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `./transcripts` | Output directory |
| `--since` | — | Only fetch notes updated after this date |
| `--dry-run` | `false` | Show what would be written, without writing |
| `--timeout` | `2m` | HTTP timeout |

### Global

| Flag | Description |
|------|-------------|
| `--debug` | Verbose logging |
| `--config` | Config file path (default: `$HOME/.granola.toml`) |

## Troubleshooting

**"No API credentials configured"** — Set `GRANOLA_API_KEY` in your environment or `.env` file.

**No transcript content** — Transcripts are only available for meetings with audio recording enabled.

**Permission denied on output directory** — Make sure the directory is writable; `sudo` is not needed.

Run `granola notes --help` for the full flag list, or add `--debug` to see detailed request logs.

---

## Contributing

Issues and PRs are welcome. To build from source:

```bash
git clone https://github.com/pradeep-stellar/granola.git
cd granola
go build -o granola .
go test ./...
```

Releases are automated via [GoReleaser](https://goreleaser.com/) — push a version tag to trigger a build.

## License

MIT — see [LICENSE](LICENSE).

## Acknowledgement

This work is based on [Christopher Lamm](https://github.com/theantichris)'s  [Granola](https://github.com/theantichris/granola).
