# Gumlet CLI

Command-line interface for managing [Gumlet](https://www.gumlet.com) video and image assets.

---

## Installation

### Homebrew (macOS / Linux)

```sh
brew install --cask gumlet/tap/gumlet
```

### npm (macOS / Linux / Windows)

```sh
npm install -g @gumlet/cli
```

### Direct download (macOS / Linux / Windows)

Download the latest binary for your platform from the [GitHub Releases](https://github.com/gumlet/cli/releases) page.

| Platform | File |
|---|---|
| macOS (Apple Silicon) | `gumlet_darwin_arm64.tar.gz` |
| macOS (Intel) | `gumlet_darwin_x86_64.tar.gz` |
| Linux (x86_64) | `gumlet_linux_x86_64.tar.gz` |
| Linux (ARM64) | `gumlet_linux_arm64.tar.gz` |
| Windows (x86_64) | `gumlet_windows_x86_64.zip` |

Extract and move to a directory on your `PATH`:

```sh
# macOS / Linux example
tar -xzf gumlet_darwin_arm64.tar.gz
sudo mv gumlet /usr/local/bin/
```

> **macOS Gatekeeper note:** If macOS blocks the binary with _"cannot be opened because the developer cannot be verified"_, remove the quarantine attribute:
> ```sh
> xattr -d com.apple.quarantine /usr/local/bin/gumlet
> ```
> Alternatively, open **System Settings → Privacy & Security** and click **Allow Anyway**. This is only needed for direct downloads — the Homebrew and npm installs handle this automatically.

---

## Authentication

Before using any command, log in with your Gumlet API key:

```sh
gumlet login
```

You will be prompted to enter your API key securely. It is saved to `~/.gumlet.yaml`.

To remove saved credentials:

```sh
gumlet logout
```

---

## Global Flags

| Flag | Default | Description |
|---|---|---|
| `--output` | `json` | Output format: `json` or `table` |

---

## Commands

### Video

#### Workspaces

| Command | Description |
|---|---|
| `gumlet video workspace list` | List all video workspaces |
| `gumlet video workspace get --workspace-id <id>` | Get details of a workspace |
| `gumlet video workspace create --name <name>` | Create a new workspace |
| `gumlet video workspace update --workspace-id <id> --name <name>` | Rename a workspace |
| `gumlet video workspace delete --workspace-id <id>` | Delete a workspace |

**Examples:**

```sh
gumlet video workspace list --output table
gumlet video workspace create --name "My Workspace"
gumlet video workspace update --workspace-id ws_123 --name "Renamed"
gumlet video workspace delete --workspace-id ws_123
```

---

#### Assets

| Command | Description |
|---|---|
| `gumlet video asset list --workspace-id <id>` | List assets in a workspace |
| `gumlet video asset get --asset-id <id>` | Get details of an asset |
| `gumlet video asset delete --asset-id <id>` | Delete an asset |

**`asset list` flags:**

| Flag | Description |
|---|---|
| `--workspace-id` | *(required)* Workspace to list assets from |
| `--status` | Filter by status (e.g. `ready`, `processing`) |
| `--tag` | Filter by tag |
| `--title` | Filter by title |
| `--folder` | Filter by folder |
| `--offset` | Pagination offset |
| `--size` | Page size |
| `--sort-by` | Field to sort by |
| `--order-by` | Sort direction (`asc` / `desc`) |

**Examples:**

```sh
gumlet video asset list --workspace-id ws_123 --output table
gumlet video asset list --workspace-id ws_123 --status ready --size 20
gumlet video asset get --asset-id asset_456
gumlet video asset delete --asset-id asset_456
```

---

#### Playlists

| Command | Description |
|---|---|
| `gumlet video playlist list` | List all playlists |
| `gumlet video playlist create --workspace-id <id> --title <title>` | Create a new playlist |
| `gumlet video playlist update --playlist-id <id>` | Update a playlist |
| `gumlet video playlist get-assets --playlist-id <id>` | Get assets in a playlist |
| `gumlet video playlist add-asset --playlist-id <id> --asset-ids <ids>` | Add assets to a playlist |
| `gumlet video playlist remove-asset --playlist-id <id> --asset-ids <ids>` | Remove assets from a playlist |

**`playlist create` flags:**

| Flag | Description |
|---|---|
| `--workspace-id` | *(required)* Workspace (collection) ID |
| `--title` | *(required)* Playlist title |
| `--description` | Playlist description |

**`playlist update` flags:**

| Flag | Description |
|---|---|
| `--playlist-id` | *(required)* Playlist ID |
| `--title` | New title |
| `--description` | New description |
| `--channel-visibility` | Channel visibility setting |

**`playlist add-asset` / `remove-asset` flags:**

| Flag | Description |
|---|---|
| `--playlist-id` | *(required)* Playlist ID |
| `--asset-ids` | *(required)* Comma-separated asset IDs |

**Examples:**

```sh
gumlet video playlist list --output table
gumlet video playlist create --workspace-id ws_123 --title "Best of 2026" --description "Top videos"
gumlet video playlist update --playlist-id pl_789 --title "Renamed Playlist"
gumlet video playlist add-asset --playlist-id pl_789 --asset-ids asset_1,asset_2
gumlet video playlist remove-asset --playlist-id pl_789 --asset-ids asset_1
gumlet video playlist get-assets --playlist-id pl_789 --output table
```

---

### Image

#### Sources

| Command | Description |
|---|---|
| `gumlet image source list` | List all image sources |
| `gumlet image source add --namespace <ns> --type <type>` | Create a new image source |
| `gumlet image source update --source-id <id>` | Update an image source |
| `gumlet image source delete --source-id <id>` | Delete an image source |

**`source add` / `update` flags:**

| Flag | Description |
|---|---|
| `--namespace` | *(add only, required)* Subdomain for the source (e.g. `mycompany`) |
| `--type` | Source type: `amazon`, `proxy`, `gcs`, `dostorage`, `wasabi`, `cloudinary`, `azure`, `linode`, `backblaze`, `cloudflare` |
| `--source-id` | *(update only, required)* ID of the source to update |
| `--config` | JSON config for the source type (see examples below) |

**Examples:**

```sh
gumlet image source list --output table

# Amazon S3
gumlet image source add \
  --namespace mycompany \
  --type amazon \
  --config '{"bucket_name":"my-bucket","bucket_region":"us-east-1","access_key":"KEY","secret":"SECRET"}'

# Web proxy
gumlet image source add \
  --namespace mycompany \
  --type proxy \
  --config '{"whitelisted_domains":["example.com"]}'

gumlet image source update --source-id src_123 --type amazon --config '{"bucket_name":"new-bucket"}'
gumlet image source delete --source-id src_123
```

---

#### Purge Cache

```sh
gumlet image purge --subdomain <subdomain> [--urls <url1,url2>]
```

| Flag | Description |
|---|---|
| `--subdomain` | *(required)* Subdomain to purge cache for |
| `--urls` | Comma-separated list of specific URLs to purge (omit to purge all) |

**Examples:**

```sh
gumlet image purge --subdomain mycompany
gumlet image purge --subdomain mycompany --urls "https://mycompany.gumlet.io/photo.jpg,https://mycompany.gumlet.io/banner.png"
```

---

## Output Formats

All commands support `--output json` (default) and `--output table`:

```sh
gumlet video workspace list --output table
gumlet video asset list --workspace-id ws_123 --output table
gumlet image source list --output table
```

---

## License

Apache 2.0 © [Gumlet Pte. Ltd.](https://www.gumlet.com)
