# Gumlet CLI

Manage [Gumlet](https://www.gumlet.com) video workspaces, assets, playlists, image sources, and cache purges from your terminal.

## Get started

### 1. Install

**Homebrew** (macOS):

```sh
brew install --cask gumlet/tap/gumlet
```

**npm** (macOS, Linux, Windows):

```sh
npm install -g @gumlet/cli
```

The npm installer downloads and verifies the matching GitHub release archive.
macOS and Linux support x86-64 and ARM64; Windows supports x86-64.

**Manual download:** Pick the archive for your OS and architecture from
[GitHub Releases](https://github.com/gumlet/cli/releases). Compare its SHA-256
digest with `checksums.txt`, extract it, and put `gumlet` (or `gumlet.exe`) on
your `PATH`. Release archives label x86-64 as `x86_64` and ARM64 as `arm64`.

Check the installation with `gumlet --version`.

### 2. Create and save an API key

Open [API Keys in the Gumlet dashboard](https://dash.gumlet.com/developer/api-keys),
generate a key, and paste it into the hidden prompt:

```sh
gumlet login
```

The CLI saves the key to `~/.gumlet.yaml` (in your user profile on Windows).
Login saves the key; the first API command verifies that it works. Rerun
`gumlet login` to replace it, or use `gumlet logout` to remove the saved key.

For CI or other non-interactive sessions, inject `GUMLET_API_KEY` from your
secret store. It takes precedence over the saved key and is not written to
disk. `gumlet logout` does not unset it.

### 3. Run your first command

```sh
gumlet video workspace list -o table
gumlet image source list -o table
```

Find the workspace or source ID in the output. Run `gumlet --help` or append
`--help` to any command for the full flag list.

## Common tasks

### Work with videos

Replace `WORKSPACE_ID`, `ASSET_ID`, and `PLAYLIST_ID` with the IDs from your account.

```sh
gumlet video asset list --workspace-id WORKSPACE_ID --status ready -o table
gumlet video asset upload --workspace-id WORKSPACE_ID --file ./intro.mp4 \
  --title "Introduction"
gumlet video asset get --asset-id ASSET_ID -o table
gumlet video playlist create --workspace-id WORKSPACE_ID --title "Getting started"
gumlet video playlist add-asset --playlist-id PLAYLIST_ID --asset-ids ASSET_ID
gumlet video playlist update --playlist-id PLAYLIST_ID --channel-visibility=false
```

Upload accepts a nonempty local file. It defaults to `ABR` output and uses the
filename as the title unless overridden. A successful upload returns
`{"asset_id":"…","uploaded":true}`. It confirms file transfer, not completed
transcoding; use `asset get` to check processing status.

### Filter and paginate asset lists

```sh
gumlet video asset list --workspace-id WORKSPACE_ID \
  --playlist-id PLAYLIST_ID --sort-by created_at --order-by desc --page-size 25

gumlet video asset list --workspace-id WORKSPACE_ID --page-size 25 --page-number 2
```

Page numbers start at 1, and page size is 1–100. For cursor pagination, pass
the response's `next_page_token` as `--next-page-token` instead of `--page-number`,
keeping the same filters and page size. Cursor pagination requires the default
or `created_at` sort and no `--title` filter. An empty token means there is no
next page. Use `--folder-id ID` to filter by folder or `--folder-id=` for
unfiled assets.

### Organize, update, and recover videos

```sh
gumlet video workspace folder create --workspace-id WORKSPACE_ID --name "Clips"
gumlet video workspace folder list --workspace-id WORKSPACE_ID -o table
gumlet video workspace folder update --workspace-id WORKSPACE_ID \
  --folder-id FOLDER_ID --asset-ids ASSET_ID
gumlet video asset update --asset-id ASSET_ID --title "New title" \
  --metadata-file ./metadata.json
gumlet video asset bulk-tag --workspace-id WORKSPACE_ID \
  --asset-ids ASSET_ID_1,ASSET_ID_2 --add-tags featured
gumlet video asset list-recoverable --workspace-id WORKSPACE_ID -o table
gumlet video asset recover --asset-id DELETED_ASSET_ID
gumlet video playlist reorder --playlist-id PLAYLIST_ID --sort-by title --sort-order asc
```

`asset update` changes only the supplied fields. Metadata must be a JSON object;
`--call-to-actions-file` takes a JSON array of objects. Either file flag accepts
`-` to read from stdin. Use `video workspace folder remove-assets` to unfile
assets without deleting them. `video asset bulk-delete` deletes the explicitly
listed assets; double-check the workspace and IDs before running it.

To move one asset in a playlist instead of sorting the whole playlist, supply
`--asset-id`, `--page-number`, `--page-size`, and `--asset-position` together.
Page numbers start at 1; positions within a page start at 0.

### Manage channel viewers

```sh
gumlet video channel viewer list --workspace-id WORKSPACE_ID --page-size 25
gumlet video channel viewer invite --workspace-id WORKSPACE_ID --viewers-file ./viewers.json
gumlet video channel viewer remove --workspace-id WORKSPACE_ID --emails viewer@example.com
```

The invite file is a JSON array such as
`[{"email":"viewer@example.com","name":"Example Viewer"}]`. Use `--viewers-file -`
to read it from stdin. List viewers by page number or pass the previous JSON
response's `next_page_token` as `--next-page-token`.

### Manage live streams

```sh
gumlet video live workspace create --name "Studio"
gumlet video live asset create --live-workspace-id LIVE_WORKSPACE_ID \
  --resolution 720p --title "Live demo"
gumlet video live asset start --live-asset-id LIVE_ASSET_ID
gumlet video live asset get --live-asset-id LIVE_ASSET_ID -o table
gumlet video live asset complete --live-asset-id LIVE_ASSET_ID
```

Live JSON responses can contain stream keys, ingest URLs, or presigned upload
URLs. Keep them out of shared logs; live table output omits those fields.
`video live asset thumbnail-upload-urls` returns upload URLs for the requested
thumbnail states in JSON but does not upload files.

### Inspect your account and invoices

```sh
gumlet user get
gumlet billing invoice history
gumlet billing invoice upcoming
```

### Manage image sources and cache

When a provider needs credentials, put its JSON configuration in a private
file or stream it from stdin. Avoid including secrets in `--config` or
committing them to your repository.

```sh
gumlet image source add --namespace mycompany --type amazon \
  --config-file ./source-config.json
gumlet image source update --source-id SOURCE_ID --type amazon \
  --config-file - < ./source-config.json

gumlet image purge --source-id SOURCE_ID --urls "/photo.jpg,/banner.png"
```

Image purge now requires `--source-id` instead of `--subdomain`. Get the ID
from `gumlet image source list -o table`. Supply explicit image paths without
query strings or fragments; purging an entire source by omitting `--urls`
is no longer supported.

## Output and scripting

API commands default to `--output json`. JSON preserves the API's field names,
nested values, and pagination metadata; an empty response writes nothing.
Choose `-o table` for a concise view with fields selected for each command.
Lists show one row per item, including one row per invoice in
`gumlet billing invoice history -o table`. Tables indicate when the API reports
more results; use JSON to retrieve cursors, full details, or links omitted from
tables. Successful mutations may print a short confirmation even when the API
returns no body. Use `--no-color` to disable styling.
Upload progress and command errors go to stderr. Command errors exit nonzero.

For scripts migrating from older versions: read nested JSON fields instead of
flattened labels, use `--page-size` and `--page-number` or `--next-page-token`
instead of `--size` / `--offset`, and use `--folder-id` instead of `--folder`.
The asset list no longer accepts `--type`. Use `--channel-visibility=true`
or `--channel-visibility=false` for playlist updates.

## Troubleshooting

- **`gumlet` not found:** Make sure the install location is on your `PATH`.
  For npm, check the global npm prefix with `npm prefix -g`.
- **Missing key or HTTP 401/403:** Check whether `GUMLET_API_KEY` is set and
  overriding your saved key. Generate a new key if needed and confirm it has
  access to the workspace or source.
- **Upload failed after creating an asset:** The error includes its asset ID.
  Check the asset before retrying; a new upload command may create another one.
- **npm installation failed:** Confirm your Node.js version meets the package's
  `engines` requirement, your platform is supported, and the machine can reach
  GitHub Releases. You can also install the release binary manually.

## AI agent skills

The repository includes skills for workspaces, assets, uploads, playlists,
live streams, channel viewers, account details and invoices, image sources,
and cache purges. Browse them in [skills](./skills) or on
[skills.sh](https://skills.sh/gumlet/cli). For example:

```sh
npx skills add https://github.com/gumlet/cli --skill gumlet-video-asset
```

## Development

Use the Go version in [go.mod](go.mod):

```sh
go run . --help
go test ./...
go vet ./...
```

The integration tests use local mock servers; no Gumlet API key is required.

To test the npm installer, use Node.js 26 (the CI version) and the same Go
toolchain. The smoke tests pack and install the package using local release
fixtures:

```sh
npm --prefix npm ci --ignore-scripts
npm --prefix npm test
```

## License

[Apache 2.0](LICENSE) © [Gumlet Pte. Ltd.](https://www.gumlet.com)
