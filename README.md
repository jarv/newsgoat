# NewsGoat

<p align="center">
  <img src="./.github/screenshot.png" alt="newsgoat screenshot" height="200">
  <img src="./.github/newsgoat.png" alt="newsgoat icon" height="200">
</p>

NewsGoat is a terminal-based RSS reader written in Go using the [bubbletea TUI framework](https://github.com/charmbracelet/bubbletea).
It's inspired by [Newsbeuter](https://github.com/akrennmair/newsbeuter)/[Newsboat](https://github.com/newsboat/newsboat) and provides a vi-like interface for reading RSS feeds.

## Why create another terminal-based RSS reader?

I’ve been using terminal-based RSS readers for about 15 years.
The first program I used was [Newsbeuter](https://github.com/akrennmair/newsbeuter), but around 2017 its maintainer announced it would no longer be maintained, so I switched to its fork, [Newsboat](https://github.com/newsboat/newsboat).
Over time, I grew frustrated with frequent crashes and started looking for alternatives. A similar CLI RSS reader written in Go called [nom](https://github.com/guyfedwards/nom) looked interesting, but it didn’t offer the feed organization I wanted.

Meanwhile, “vibe coding” was catching on, and it seemed like a fun excuse to see how quickly I could build my own news reader in Go with exactly the features I wanted. After about a day of prompting, tweaking, and vibing through the process, here’s the result—enjoy!

## Features

- **Folder organization**: Organize feeds into collapsible folders. Feeds can belong to multiple folders.
- **Feed grouping**: Feeds are grouped by the feed URL.
- **Feed info**: Quickly query info for every feed including cache-control, last-updated, etc. Press <kbd>i</kbd> to view feed info.
- **Error logs**: Error logs are shown directly in the UI for feed troubleshooting. Press <kbd>l</kbd> to view logs.
- **Task management**: Refresh task control separate in the app with a way to see what is queued, running and failures. Press <kbd>t</kbd> to view tasks.
- **Flexible sorting**: Option to put feeds with unread items at the top. Press <kbd>c</kbd> to configure.
- **Auto-discovery**: Automatic feed discovery when adding URLs. Press <kbd>u</kbd> to add a youtube link and automatically subscribe to the channel's feed.

## Feed Auto Discovery

### Watch individual files or directories in a GitHub/GitLab repository

It is possible to monitor individual files (or directories) in GitHub/GitLab by subscribing to that files commit history.
There is first-class support for this by parsing any GitHub/GitLab link when you press <kbd>u</kbd> to add a URL and will subscribe to the feed.
The name of the feed will be displayed as the path to the file.
If `GITHUB_FEED_TOKEN` or `GITLAB_FEED_TOKEN` is set in the environment, it will use that as part of fetch for private repositories.

### Youtube

Subscribe to a YouTube channel with RSS by pressing <kbd>u</kbd> to add a YouTube URL.
This will extract the `channel_id` and subscribe to the channel RSS feed.

## Design Principles

- **Beautiful and compact**: Compact design and tactful use of emojis.
- **Opinionated**: It was built with one (my own) preferred configuration for Newsboat in mind, it is not as configurable as the alternatives.
- **A good "netizen"**: It follows [rachelbythebay](https://rachelbythebay.com/) [feed reader best practices](https://rachelbythebay.com/fs/help.html).
  - sends conditional responses
  - respects `cache-control` and will use the local cache instead instead of a conditional response
  - sets a useful user-agent
- **Local only**: There are no current plans for cloud syncing, sorry!
- **URLs as plain text**: I am not a fan of yaml based configuration so feed URLs are in a plain text file similar to Newsboat
- **Configuration in the UI**: For what little configuration there is, it is set in the UI instead of through a configuration file

## Alternatives

- The original CLI based newsreader [newsbeuter](https://github.com/akrennmair/newsbeuter).
- Newsbeuter was archived, and I think was forked as [newsboat](https://github.com/newsboat/newsboat) and re-written in Rust.
- [nom](https://github.com/guyfedwards/nom) is a similar CLI based news reader (also written in Go).

If you know of any other CLI based RSS readers worth mentioning here please add them!

## Configure

Create a `.config/newsgoat/urls` file with one feed per line.

## Development

### Setup

1. Install [mise](https://mise.jdx.dev/) for tool version management
2. Clone the repository
3. Enable git hooks:

```bash
mise run setup-hooks
```

This will configure git to run the linter before each commit.

### Building

```bash
mise run build    # Build the binary
go run .          # Run with urls file in .config/newsgoat/urls
go run . -urlFile urls.example  # Run using the example urls file
```

### Linting

The project uses golangci-lint for code quality:

```bash
mise run lint     # Run the linter
```

The linter runs automatically on every commit via a pre-commit hook. To bypass temporarily (not recommended):

```bash
git commit --no-verify
```

## Install

### Quick Install (Recommended)

Install with a single command (macOS and Linux):

```bash
curl -sSL https://raw.githubusercontent.com/jarv/newsgoat/main/install.sh | bash
```

This will automatically detect your OS and architecture and install the latest version to `/usr/local/bin`.

### Manual Install

Download the latest release for your platform from the [releases page](https://github.com/jarv/newsgoat/releases):

#### macOS (Apple Silicon / Intel)

<details>

**Apple Silicon**

```bash
# Apple Silicon
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-darwin-arm64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

**Intel**

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-darwin-amd64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

</details>

#### Linux (amd64 / arm64)

<details>

**amd64**

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-linux-amd64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

**arm64**

```bash
curl -L https://github.com/jarv/newsgoat/releases/latest/download/newsgoat-linux-arm64 -o newsgoat
chmod +x newsgoat
sudo mv newsgoat /usr/local/bin/
```

</details>

## Add Feed URLs

There are three ways to add feed URLs to NewsGoat:

### 1. Via Command Line

Add a feed URL directly from the command line with automatic feed discovery:

```bash
newsgoat add <url>
```

NewsGoat will automatically discover the RSS/Atom feed URL from the provided URL. For example:

- `newsgoat add https://example.com` will find the feed link in the page
- `newsgoat add https://youtube.com/@channel` will discover the YouTube RSS feed

### 2. In the Application (Interactive)

Press `u` in the feed list view to open an interactive prompt where you can:

- Type or paste a URL (optionally followed by folders)
- Format: `<url>` or `<url> folder1,folder2` or `<url> "folder with spaces",folder3`
- Press Enter to add (with automatic feed discovery)
- Press Esc to cancel

Examples:

- `https://example.com Tech News`
- `https://youtube.com/@channel Tech News,YouTube`
- `https://example.com "My Folder",Tech`

### 3. Edit the URLs File Directly

Press `U` (Shift+U) in the feed list view to open `~/.config/newsgoat/urls` in your `$EDITOR`.

Alternatively, edit the file manually:

- Create/edit `~/.config/newsgoat/urls`
- Add one feed URL per line
- Optionally add folders after the URL: `<url> folder1,folder2`
- Use quotes for folder names with spaces: `<url> "folder name",otherfolder`
- Lines starting with `#` are treated as comments
- Save and press `Ctrl+R` in NewsGoat to reload

Example `urls` file:

```text
# Feeds can have folders! Format: <url> folder1,folder2
# Use quotes for folder names with spaces: <url> "folder name",otherfolder

# Tech News
https://www.theverge.com/rss/index.xml Tech News
https://feeds.feedburner.com/TechCrunch/ Tech News,Startups
https://www.wired.com/feed/rss Tech News

# Science feeds
https://www.sciencedaily.com/rss/top/science.xml Science
https://www.nasa.gov/rss/dyn/breaking_news.rss Science,Space

# Feeds with spaces in folder names
https://www.popularmechanics.com/rss/all.xml/ "DIY & Tech"

# Feeds without folders
https://example.com/feed
```

## Organizing Feeds with Folders

NewsGoat supports organizing feeds into folders:

- **Multiple folders**: Feeds can belong to multiple folders and will appear under each one
- **Collapsible**: Press Enter on a folder to expand/collapse its contents
- **Visual hierarchy**: Feeds under folders are displayed with a vertical bar (`│`) for easy identification
- **Folder operations**:
  - Press `r` on a folder to refresh all feeds in that folder
  - Press `A` on a folder to mark all items in that folder as read
- **Sorting**: When "Unread on Top" is enabled:
  - Unread feeds without folders appear at the very top
  - Within folders, unread feeds appear before read feeds

## Searching Feeds and Articles

NewsGoat provides two search modes with case-insensitive text matching:

### Search Modes

- **Global Search** (<kbd>/</kbd>): Searches across multiple fields
  - **Feed List View**: Searches all feed content
  - **Item List View**: Searches all feed content across feed items

- **Title Search** (<kbd>Ctrl</kbd>+<kbd>F</kbd>): Searches only titles
  - **Feed List View**: Searches feed titles only
  - **Item List View**: Searches item titles only

### Using Search

1. Press <kbd>/</kbd> or <kbd>Ctrl</kbd>+<kbd>F</kbd> to start searching
2. Type your search query (case-insensitive text matching)
3. Switch between modes anytime by pressing <kbd>/</kbd> or <kbd>Ctrl</kbd>+<kbd>F</kbd>
4. Press <kbd>Enter</kbd> to exit search mode and keep results
5. Press <kbd>Esc</kbd> to cancel and restore original view

Search filters results in real-time as you type, making it easy to find specific feeds or articles quickly.

## Keys

### Global (Available in All Views)

| Key | Description |
|-----|-------------|
| <kbd>h</kbd>, <kbd>?</kbd> | Show help |
| <kbd>q</kbd> | Go back / quit (press twice in feed view) |
| <kbd>Esc</kbd> | Go back (does nothing in feed view) |
| <kbd>Ctrl</kbd>+<kbd>C</kbd> | Go back / quit (press twice in feed view) |
| <kbd>j</kbd>, <kbd>↓</kbd> | Move down |
| <kbd>k</kbd>, <kbd>↑</kbd> | Move up |
| <kbd>Enter</kbd> | Select / open |
| <kbd>Ctrl</kbd>+<kbd>D</kbd> | Page down |
| <kbd>Ctrl</kbd>+<kbd>U</kbd> | Page up |

### Feed List View

| Key | Description |
|-----|-------------|
| <kbd>Enter</kbd> | Open feed / expand or collapse folder |
| <kbd>r</kbd> | Refresh selected feed or all feeds in folder |
| <kbd>R</kbd> | Refresh all feeds |
| <kbd>A</kbd> | Mark all items in feed/folder as read |
| <kbd>i</kbd> | Show feed info (cache-control, last-updated, etc.) |
| <kbd>/</kbd> | Global search (all feed content) |
| <kbd>Ctrl</kbd>+<kbd>F</kbd> | Title search only |
| <kbd>u</kbd> | Add URL with optional folders (e.g., `url folder1,folder2`) |
| <kbd>U</kbd> | Edit URLs file in $EDITOR |
| <kbd>Ctrl</kbd>+<kbd>R</kbd> | Reload URLs from file |
| <kbd>l</kbd> | View logs |
| <kbd>t</kbd> | View tasks |
| <kbd>c</kbd> | View settings |

### Item List View (Articles in a Feed)

| Key | Description |
|-----|-------------|
| <kbd>/</kbd> | Global search (all feed content) |
| <kbd>Ctrl</kbd>+<kbd>F</kbd> | Title search only |
| <kbd>r</kbd> | Refresh current feed |
| <kbd>R</kbd> | Refresh all feeds |
| <kbd>A</kbd> | Mark all items as read |
| <kbd>N</kbd> | Toggle read status of selected item |
| <kbd>o</kbd> | Open item link in browser |
| <kbd>c</kbd> | View settings |
| <kbd>t</kbd> | View tasks |

### Article View

| Key | Description |
|-----|-------------|
| <kbd>1-9</kbd> | Open numbered link in browser |
| <kbd>o</kbd> | Open article link in browser |
| <kbd>n</kbd> | Next article |
| <kbd>N</kbd> | Previous article |
| <kbd>r</kbd> | Toggle raw HTML view |
| <kbd>c</kbd> | View settings |
| <kbd>t</kbd> | View tasks |

### Tasks View

| Key | Description |
|-----|-------------|
| <kbd>d</kbd> | Remove selected task |
| <kbd>c</kbd> | Clear all failed tasks |
| <kbd>l</kbd> | View logs |

### Log View

| Key | Description |
|-----|-------------|
| <kbd>c</kbd> | Clear all log messages |

### Settings View

| Key | Description |
|-----|-------------|
| <kbd>?</kbd> | Toggle settings help |
| <kbd>Enter</kbd> | Edit selected setting |

### Status Icons

| Icon | Meaning |
|------|---------|
| 📁 | Closed folder |
| 📂 | Open folder |
| 🔵 | Unread items/feed |
| 🔍 | 404 Not Found |
| 🚫 | 403 Forbidden |
| ⏱️ | 429 Too Many Requests |
| ⚠️ | 500/502/503 Server Error |
| ⌛ | Timeout |
| ❌ | Other Error |
| 🕓 | Pending task |
| 🔄 | Running task |
| 💥 | Failed task |
| │ | Feed under folder (vertical bar prefix) |
