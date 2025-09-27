# NewsGoat

<img src="./.github/newsgoat.png" width="200">

NewsGoat is a terminal-based RSS reader written in Go using the [bubbletea TUI framework](https://github.com/charmbracelet/bubbletea).
It's inspired by [newsboat](https://github.com/newsboat/newsboat) and provides a vi-like interface for reading RSS feeds.

## Design Principles

- **Opinionated**: It was built with my preferred configuration for Newsboat in mind.
- **A good "netizen"**: It follows [feed reader best practices](https://rachelbythebay.com/fs/help.html), including
  - respecting cache-control sent by the feed server
  - sending conditional responses
  - setting a useful user-agent
- **Local only**: There are no current plans for cloud syncing, sorry!
- **URLs as plain text**: I am not a fan of yaml based configuration so feed URLs are in a plain text file similar to Newsboat
- **Configuration in the UI**: For what little configuraiton there is, it is set in the UI instead of through a configuration file

## Alternatives

- The original CLI based newsreader [newsbeuter](https://github.com/akrennmair/newsbeuter).
- Newsbeuter was archived, and I think was forked as [newsboat](https://github.com/newsboat/newsboat) and re-written in Rust.
- [nom](https://github.com/guyfedwards/nom) is a similar CLI based news reader (also written in Go).

If you know of any other CLI based RSS readers worth mentioning here please add them!

## Build and Run

```bash
mise install
mise run build          # Build binary
mise run run            # Run the application directly
mise run watch          # Watch for changes and auto-run
mise run golangci-lint  # Run linting
```

## Install

### From Release

Download the latest release for your platform from the [releases page](https://github.com/jarv/newsgoat/releases):

#### macOS (Apple Silicon)

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-darwin-arm64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

#### macOS (Intel)

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-darwin-amd64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

#### Linux (amd64)

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-linux-amd64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

#### Linux (arm64)

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-linux-arm64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```
