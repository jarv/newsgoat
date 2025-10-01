# NewsGoat

<p align="left">
  <img src="./.github/screenshot.png" height="200">
  <img src="./.github/newsgoat.png" height="200">
</p>

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

## Configure

Create a `.config/newsgoat/urls` file with one feed per line.

## Build and Run

```bash
go run . # Run with urls file in .config/newsgoat/urls
go run . -urlFile urls.example # Run using the example urls file
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
