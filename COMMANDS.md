# Gumlet CLI Command Reference

Complete reference for all commands, flags, arguments, and output conventions in the Gumlet CLI.

## Table of Contents

- [Global Flags & Conventions](#global-flags--conventions)
- [Core Commands](#core-commands)
  - [`gumlet login`](#gumlet-login)
  - [`gumlet logout`](#gumlet-logout)
  - [`gumlet version`](#gumlet-version)
- [Account & Billing](#account--billing)
  - [`gumlet user get`](#gumlet-user-get)
  - [`gumlet org get`](#gumlet-org-get)
  - [`gumlet audit-log list`](#gumlet-audit-log-list)
  - [`gumlet billing invoice history`](#gumlet-billing-invoice-history)
  - [`gumlet billing invoice upcoming`](#gumlet-billing-invoice-upcoming)
- [Image Sources & Cache](#image-sources--cache)
  - [`gumlet image purge`](#gumlet-image-purge)
  - [`gumlet image source list`](#gumlet-image-source-list)
  - [`gumlet image source get`](#gumlet-image-source-get)
  - [`gumlet image source add`](#gumlet-image-source-add)
  - [`gumlet image source update`](#gumlet-image-source-update)
  - [`gumlet image source delete`](#gumlet-image-source-delete)
- [Video Workspaces](#video-workspaces)
  - [`gumlet video workspace list`](#gumlet-video-workspace-list)
  - [`gumlet video workspace get`](#gumlet-video-workspace-get)
  - [`gumlet video workspace create`](#gumlet-video-workspace-create)
  - [`gumlet video workspace update`](#gumlet-video-workspace-update)
  - [`gumlet video workspace delete`](#gumlet-video-workspace-delete)
- [Video Workspace Folders](#video-workspace-folders)
  - [`gumlet video workspace folder list`](#gumlet-video-workspace-folder-list)
  - [`gumlet video workspace folder get`](#gumlet-video-workspace-folder-get)
  - [`gumlet video workspace folder create`](#gumlet-video-workspace-folder-create)
  - [`gumlet video workspace folder update`](#gumlet-video-workspace-folder-update)
  - [`gumlet video workspace folder delete`](#gumlet-video-workspace-folder-delete)
  - [`gumlet video workspace folder remove-assets`](#gumlet-video-workspace-folder-remove-assets)
- [Video Assets](#video-assets)
  - [`gumlet video asset list`](#gumlet-video-asset-list)
  - [`gumlet video asset get`](#gumlet-video-asset-get)
  - [`gumlet video asset create`](#gumlet-video-asset-create)
  - [`gumlet video asset upload`](#gumlet-video-asset-upload)
  - [`gumlet video asset update`](#gumlet-video-asset-update)
  - [`gumlet video asset delete`](#gumlet-video-asset-delete)
  - [`gumlet video asset thumbnail-upload`](#gumlet-video-asset-thumbnail-upload)
  - [`gumlet video asset thumbnail-select`](#gumlet-video-asset-thumbnail-select)
  - [`gumlet video asset bulk-delete`](#gumlet-video-asset-bulk-delete)
  - [`gumlet video asset bulk-tag`](#gumlet-video-asset-bulk-tag)
  - [`gumlet video asset list-recoverable`](#gumlet-video-asset-list-recoverable)
  - [`gumlet video asset recover`](#gumlet-video-asset-recover)
- [Video Playlists](#video-playlists)
  - [`gumlet video playlist list`](#gumlet-video-playlist-list)
  - [`gumlet video playlist create`](#gumlet-video-playlist-create)
  - [`gumlet video playlist update`](#gumlet-video-playlist-update)
  - [`gumlet video playlist delete`](#gumlet-video-playlist-delete)
  - [`gumlet video playlist get-assets`](#gumlet-video-playlist-get-assets)
  - [`gumlet video playlist add-asset`](#gumlet-video-playlist-add-asset)
  - [`gumlet video playlist remove-asset`](#gumlet-video-playlist-remove-asset)
  - [`gumlet video playlist reorder`](#gumlet-video-playlist-reorder)
- [Channel Viewers](#channel-viewers)
  - [`gumlet video channel viewer list`](#gumlet-video-channel-viewer-list)
  - [`gumlet video channel viewer invite`](#gumlet-video-channel-viewer-invite)
  - [`gumlet video channel viewer remove`](#gumlet-video-channel-viewer-remove)
- [Live Workspaces](#live-workspaces)
  - [`gumlet video live workspace list`](#gumlet-video-live-workspace-list)
  - [`gumlet video live workspace create`](#gumlet-video-live-workspace-create)
  - [`gumlet video live workspace update`](#gumlet-video-live-workspace-update)
  - [`gumlet video live workspace delete`](#gumlet-video-live-workspace-delete)
- [Live Assets](#live-assets)
  - [`gumlet video live asset list`](#gumlet-video-live-asset-list)
  - [`gumlet video live asset get`](#gumlet-video-live-asset-get)
  - [`gumlet video live asset create`](#gumlet-video-live-asset-create)
  - [`gumlet video live asset update`](#gumlet-video-live-asset-update)
  - [`gumlet video live asset delete`](#gumlet-video-live-asset-delete)
  - [`gumlet video live asset start`](#gumlet-video-live-asset-start)
  - [`gumlet video live asset complete`](#gumlet-video-live-asset-complete)
  - [`gumlet video live asset history`](#gumlet-video-live-asset-history)
  - [`gumlet video live asset thumbnail-upload-urls`](#gumlet-video-live-asset-thumbnail-upload-urls)

---

## Global Flags & Conventions

### Global Flags

Available on all commands:

| Flag | Short | Type | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--output` | `-o` | `string` | `json` | Output format: `json` or `table` |
| `--no-color` | | `bool` | `false` | Disable colored output |
| `--help` | `-h` | `bool` | `false` | Show help for any command |
| `--version` | `-v` | `bool` | `false` | Print version information (root command) |

### Conventions

- **Positional arguments:** No commands accept positional arguments; all inputs are passed via flags.
- **Authentication:** Commands read `GUMLET_API_KEY` from the environment first, falling back to `~/.gumlet.yaml` (saved via `gumlet login`).
- **Output:** API commands default to raw JSON (`-o json`) preserving API field names, nested values, and pagination cursors. Use `-o table` for concise human-readable tables. Upload progress and errors are written to stderr; errors exit nonzero.
- **List flags (`stringSlice` / `intSlice`):** Accept comma-separated values (e.g. `--asset-ids id1,id2`) or repeated flags (`--asset-ids id1 --asset-ids id2`).
- **File flags:** Flags ending in `-file` accept `-` to read from stdin (max 1 MiB).
- **Boolean flags:** Pass explicit values with `=` when setting to `false` (e.g. `--channel-visibility=false`).

---

## Core Commands

### `gumlet login`

Sign in to Gumlet via browser using OAuth 2.0 (Authorization Code + PKCE with a localhost callback) and save the resulting tokens to `~/.gumlet.yaml`. For CI or non-interactive environments, create an API key in [Gumlet API Keys](https://dash.gumlet.com/developer/api-keys) and set `GUMLET_API_KEY` instead.

| Flag | Type | Default | Description |
|---|---|---|---|
| `--no-browser` | `bool` | `false` | Print the sign-in URL instead of opening the browser automatically. |

```sh
gumlet login
gumlet login --no-browser
```

---

### `gumlet logout`

Prompt for confirmation (`y`/`N`), revoke the saved OAuth token, and remove `~/.gumlet.yaml`. Does not unset `GUMLET_API_KEY`.

*No command-specific flags.*

```sh
gumlet logout
```

---

### `gumlet version`

Print CLI version, git commit, and build timestamp.

*No command-specific flags.*

```sh
gumlet version
```

---

## Account & Billing

### `gumlet user get`

Get the current user's details.

- **API:** `GET /user/data`
- *No command-specific flags.*

```sh
gumlet user get -o table
```

---

### `gumlet org get`

Get details of the current organization (plan, billing cycle, SSO status, video/image plans).

- **API:** `GET /org/data`
- *No command-specific flags.*

```sh
gumlet org get -o table
```

---

### `gumlet audit-log list`

Fetch organization audit logs. Accessible to `owner` and `admin` role users only.

- **API:** `POST /user/audit-log`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--start-at` | `string` | Yes | | Start date in `YYYY-MM-DD` format |
| `--end-at` | `string` | Yes | | End date in `YYYY-MM-DD` format (must not be before `--start-at`) |
| `--page-number` | `int64` | No | `1` | Page number (starts at `1`) |
| `--page-size` | `int64` | No | `100` | Page size (`10`–`500`) |
| `--user-email` | `stringSlice` | No | | Comma-separated bare email addresses to filter by |
| `--activity-type` | `stringSlice` | No | | Comma-separated activity types to filter by (e.g. `video_uploaded`, `workspace_created`) |

```sh
gumlet audit-log list --start-at 2026-08-01 --end-at 2026-08-31 --page-size 50 -o table
```

---

### `gumlet billing invoice history`

List past Stripe invoices and unpaid invoice count.

- **API:** `GET /mixed/billing/invoice/history`
- *No command-specific flags.*

```sh
gumlet billing invoice history -o table
```

---

### `gumlet billing invoice upcoming`

Get the upcoming invoice and current month usage.

- **API:** `GET /mixed/billing/invoice/upcoming`
- *No command-specific flags.*

```sh
gumlet billing invoice upcoming -o table
```

---

## Image Sources & Cache

### `gumlet image purge`

Purge cached images for an image source.

- **API:** `POST /image/purge/{source_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--source-id` | `string` | Yes | | ID of the image source |
| `--urls` | `stringSlice` | Yes | | Comma-separated relative image paths without scheme, host, query string, or fragment |

```sh
gumlet image purge --source-id SOURCE_ID --urls "/photo.jpg,/banner.png"
```

---

### `gumlet image source list`

List all image sources.

- **API:** `GET /image/sources`
- *No command-specific flags.*

```sh
gumlet image source list -o table
```

---

### `gumlet image source get`

Get details of an image source.

- **API:** `GET /image/sources/{image_source_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--source-id` | `string` | Yes | | ID of the image source |

```sh
gumlet image source get --source-id SOURCE_ID -o table
```

---

### `gumlet image source add`

Create a new image source.

- **API:** `POST /image/sources`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--namespace` | `string` | Yes | | Subdomain for the image source (e.g. `mycompany`) |
| `--type` | `string` | Yes | | Provider type: `amazon`, `proxy`, `gcs`, `dostorage`, `wasabi`, `cloudinary`, `azure`, `linode`, `backblaze`, `cloudflare` |
| `--config` | `string` | No | | Inline provider configuration as a JSON object |
| `--config-file` | `string` | No | | Path to a JSON object file with provider configuration (`-` for stdin; max 1 MiB) |

**Constraints:** `--config` and `--config-file` are mutually exclusive.

```sh
gumlet image source add --namespace mycompany --type amazon --config-file ./source-config.json
```

---

### `gumlet image source update`

Update an image source.

- **API:** `POST /image/sources/{image_source_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--source-id` | `string` | Yes | | ID of the image source to update |
| `--type` | `string` | Yes | | Provider type: `amazon`, `proxy`, `gcs`, `dostorage`, `wasabi`, `cloudinary`, `azure`, `linode`, `backblaze`, `cloudflare` |
| `--config` | `string` | No | | Inline provider configuration as a JSON object |
| `--config-file` | `string` | No | | Path to a JSON object file with provider configuration (`-` for stdin; max 1 MiB) |

**Constraints:** `--config` and `--config-file` are mutually exclusive.

```sh
gumlet image source update --source-id SOURCE_ID --type amazon --config-file - < ./source-config.json
```

---

### `gumlet image source delete`

Delete an image source.

- **API:** `DELETE /image/sources/{image_source_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--source-id` | `string` | Yes | | ID of the image source to delete |

```sh
gumlet image source delete --source-id SOURCE_ID
```

---

## Video Workspaces

### `gumlet video workspace list`

List all video workspaces.

- **API:** `GET /video/workspaces`
- *No command-specific flags.*

```sh
gumlet video workspace list -o table
```

---

### `gumlet video workspace get`

Get details of a video workspace.

- **API:** `GET /video/workspaces/{workspace_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |

```sh
gumlet video workspace get --workspace-id WORKSPACE_ID -o table
```

---

### `gumlet video workspace create`

Create a new video workspace.

- **API:** `POST /video/workspaces`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--name` | `string` | Yes | | Name of the new workspace |

```sh
gumlet video workspace create --name "Production Videos"
```

---

### `gumlet video workspace update`

Update a video workspace.

- **API:** `POST /video/workspaces/{workspace_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace to update |
| `--name` | `string` | Yes | | New name for the workspace |

```sh
gumlet video workspace update --workspace-id WORKSPACE_ID --name "Renamed Workspace"
```

---

### `gumlet video workspace delete`

Delete a video workspace.

- **API:** `DELETE /video/workspaces/{workspace_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace to delete |

```sh
gumlet video workspace delete --workspace-id WORKSPACE_ID
```

---

## Video Workspace Folders

All `gumlet video workspace folder` subcommands require `--workspace-id`.

### `gumlet video workspace folder list`

List folders in a video workspace.

- **API:** `GET /video/workspaces/{workspace_id}/folders`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--parent-id` | `string` | No | | Filter by parent folder ID; use `"null"` for root folders |

```sh
gumlet video workspace folder list --workspace-id WORKSPACE_ID -o table
```

---

### `gumlet video workspace folder get`

Get details of a folder.

- **API:** `GET /video/workspaces/{workspace_id}/folders/{folder_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--folder-id` | `string` | Yes | | ID of the folder |

```sh
gumlet video workspace folder get --workspace-id WORKSPACE_ID --folder-id FOLDER_ID -o table
```

---

### `gumlet video workspace folder create`

Create a folder in a video workspace.

- **API:** `POST /video/workspaces/{workspace_id}/folders`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--name` | `string` | Yes | | Folder name |
| `--parent-id` | `string` | No | | Parent folder ID; omit to create a root folder |

```sh
gumlet video workspace folder create --workspace-id WORKSPACE_ID --name "Clips"
```

---

### `gumlet video workspace folder update`

Rename a folder, move it to another parent (or root), or move assets into it.

- **API:** `POST /video/workspaces/{workspace_id}/folders/{folder_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--folder-id` | `string` | Yes | | ID of the folder |
| `--name` | `string` | Conditional | | New folder name |
| `--parent-id` | `string` | No | | New parent folder ID |
| `--root` | `bool` | No | `false` | Move the folder to the root level |
| `--asset-ids` | `stringSlice` | Conditional | | Asset IDs to move into the folder |

**Constraints:**
- At least one of `--name` or `--asset-ids` is required.
- `--root` and `--parent-id` cannot be combined.
- `--root` requires `--name` (pass the current name to keep it unchanged).

```sh
gumlet video workspace folder update --workspace-id WORKSPACE_ID --folder-id FOLDER_ID --asset-ids ASSET_ID_1,ASSET_ID_2
```

---

### `gumlet video workspace folder delete`

Delete a folder.

- **API:** `DELETE /video/workspaces/{workspace_id}/folders/{folder_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--folder-id` | `string` | Yes | | ID of the folder |

```sh
gumlet video workspace folder delete --workspace-id WORKSPACE_ID --folder-id FOLDER_ID
```

---

### `gumlet video workspace folder remove-assets`

Remove assets from their folders without deleting the assets.

- **API:** `POST /video/workspaces/{workspace_id}/remove-assets-from-folder`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--asset-ids` | `stringSlice` | Yes | | Asset IDs to remove from folders |

```sh
gumlet video workspace folder remove-assets --workspace-id WORKSPACE_ID --asset-ids ASSET_ID_1,ASSET_ID_2
```

---

## Video Assets

### `gumlet video asset list`

List assets in a video workspace with filtering, sorting, and pagination.

- **API:** `GET /video/assets/list/{workspace_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | Workspace ID to list assets from |
| `--status` | `string` | No | | Filter by asset status (e.g. `ready`, `processing`, `errored`) |
| `--tag` | `string` | No | | Filter by asset tag |
| `--title` | `string` | No | | Filter by asset title |
| `--folder-id` | `string` | No | | Filter by folder ID; pass `--folder-id=` (empty) for unfiled assets |
| `--playlist-id` | `string` | No | | Filter by playlist ID |
| `--sort-by` | `string` | No | | Sort field: `created_at`, `uploaded_at`, `title`, `duration`, or `status` |
| `--order-by` | `string` | No | | Sort direction: `asc` or `desc` |
| `--page-number` | `int64` | No | `1` | Page number (starts at `1`) |
| `--page-size` | `int64` | No | `20` | Page size (`1`–`100`) |
| `--next-page-token` | `string` | No | | Cursor returned by the previous page |

**Constraints:**
- `--page-number` and `--next-page-token` are mutually exclusive.
- `--next-page-token` requires the default or `created_at` sort and no `--title` filter.

```sh
gumlet video asset list --workspace-id WORKSPACE_ID --status ready --sort-by created_at --order-by desc --page-size 25 -o table
```

---

### `gumlet video asset get`

Get details of a video asset.

- **API:** `GET /video/assets/{asset_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--asset-id` | `string` | Yes | | ID of the asset |

```sh
gumlet video asset get --asset-id ASSET_ID -o table
```

---

### `gumlet video asset create`

Create a video asset by ingesting from a remote URL or web address.

- **API:** `POST /video/assets`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | Workspace (collection) ID |
| `--input` | `string` | Yes | | URL or web address of the video file to ingest |
| `--format` | `string` | No | `ABR` | Transcode format: `ABR` or `MP4` (omitted when `--profile-id` is set without `--format`) |
| `--title` | `string` | No | | Asset title |
| `--description` | `string` | No | | Asset description |
| `--profile-id` | `string` | No | | Video profile ID |
| `--playlist-id` | `string` | No | | Add asset to this playlist |
| `--folder-id` | `string` | No | | Add asset to an existing folder |
| `--tag` | `stringSlice` | No | | Comma-separated tags to apply |
| `--resolution` | `string` | No | | Comma-separated ABR resolutions (`240p`, `360p`, `480p`, `540p`, `720p`, `1080p`) |
| `--mp4-access` | `bool` | No | `false` | Create a downloadable MP4 version |
| `--audio-only` | `bool` | No | `false` | Transcode and deliver audio-only |
| `--enable-drm` | `bool` | No | `false` | Enable DRM encryption |
| `--metadata-file` | `string` | No | | Read metadata from a JSON object file (`-` for stdin; max 1 MiB) |

```sh
gumlet video asset create --workspace-id WORKSPACE_ID --input https://example.com/video.mp4 --title "Demo"
```

---

### `gumlet video asset upload`

Upload a nonempty local regular video file as a new asset. Creates a direct upload session and uploads the file via HTTP PUT.

- **API:** `POST /video/assets/upload` (followed by direct upload PUT)

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--file` | `string` | Yes | | Path to the local video file to upload |
| `--workspace-id` | `string` | Yes | | Workspace (collection) ID |
| `--format` | `string` | No | `ABR` | Transcode format: `ABR` or `MP4` |
| `--title` | `string` | No | *filename* | Asset title (defaults to the file's basename) |
| `--description` | `string` | No | | Asset description |
| `--profile-id` | `string` | No | | Video profile ID |
| `--playlist-id` | `string` | No | | Add asset to this playlist |
| `--tag` | `stringSlice` | No | | Comma-separated tags to apply |

**Output:** Returns `{"asset_id":"...","uploaded":true}` in JSON mode once the transfer completes (transcoding continues asynchronously; use `gumlet video asset get` to check status). Progress is printed to stderr when attached to a TTY.

```sh
gumlet video asset upload --workspace-id WORKSPACE_ID --file ./intro.mp4 --title "Introduction" --tag onboarding,public
```

---

### `gumlet video asset update`

Update a video asset's metadata or processing settings. Only supplied fields are modified.

- **API:** `POST /video/assets/update`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--asset-id` | `string` | Yes | | ID of the asset |
| `--title` | `string` | No | | New title |
| `--description` | `string` | No | | New description |
| `--tag` | `string` | No | | New comma-separated tags |
| `--input` | `string` | No | | Replacement input path in Gumlet storage |
| `--reprocess` | `bool` | No | `false` | Reprocess the video |
| `--remove-subtitles` | `stringSlice` | No | | Subtitle language codes to remove |
| `--metadata-file` | `string` | No | | Read metadata from a JSON object file (`-` for stdin; max 1 MiB) |
| `--call-to-actions-file` | `string` | No | | Read calls to action from a JSON array of objects file (`-` for stdin; max 1 MiB) |

**Constraints:**
- At least one update flag (`--title`, `--description`, `--tag`, `--input`, `--reprocess`, `--remove-subtitles`, `--metadata-file`, or `--call-to-actions-file`) is required.
- `--metadata-file` and `--call-to-actions-file` cannot both read from stdin (`-`).

```sh
gumlet video asset update --asset-id ASSET_ID --title "New Title" --metadata-file ./metadata.json
```

---

### `gumlet video asset delete`

Delete a video asset.

- **API:** `DELETE /video/assets/{asset_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--asset-id` | `string` | Yes | | ID of the asset to delete |

```sh
gumlet video asset delete --asset-id ASSET_ID
```

---

### `gumlet video asset thumbnail-upload`

Upload a local image file as an asset's custom thumbnail. Requests a presigned upload URL and uploads the file via HTTP PUT.

- **API:** `POST /video/assets/{asset_ID}/thumbnail` (followed by direct upload PUT)

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--asset-id` | `string` | Yes | | ID of the video asset |
| `--file` | `string` | Yes | | Path to the local image file to upload |

**Output:** Returns `{"asset_id":"...","uploaded":true,"thumbnail_updated_at":...}` in JSON mode once the upload completes.

```sh
gumlet video asset thumbnail-upload --asset-id ASSET_ID --file ./thumb.png
```

---

### `gumlet video asset thumbnail-select`

Select a frame from a video asset to use as its thumbnail.

- **API:** `POST /video/assets/{asset_id}/thumbnail-select`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--asset-id` | `string` | Yes | | ID of the video asset |
| `--frame-at-second` | `float64` | Yes | | Non-negative timestamp in seconds (fractional values supported) |

```sh
gumlet video asset thumbnail-select --asset-id ASSET_ID --frame-at-second 2.5
```

---

### `gumlet video asset bulk-delete`

Delete multiple video assets in a workspace.

- **API:** `DELETE /video/assets/bulk/delete`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--asset-ids` | `stringSlice` | Yes | | Asset IDs to delete |

```sh
gumlet video asset bulk-delete --workspace-id WORKSPACE_ID --asset-ids ASSET_ID_1,ASSET_ID_2
```

---

### `gumlet video asset bulk-tag`

Add or remove tags on multiple video assets.

- **API:** `POST /video/assets/bulk/tag`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--asset-ids` | `stringSlice` | Yes | | Asset IDs to modify |
| `--add-tags` | `stringSlice` | Conditional | | Tags to add |
| `--remove-tags` | `stringSlice` | Conditional | | Tags to remove |

**Constraints:** At least one of `--add-tags` or `--remove-tags` must be provided.

```sh
gumlet video asset bulk-tag --workspace-id WORKSPACE_ID --asset-ids ASSET_ID_1,ASSET_ID_2 --add-tags featured --remove-tags draft
```

---

### `gumlet video asset list-recoverable`

List deleted assets in a workspace's recycle bin.

- **API:** `GET /video/asset/recoverable/list`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the workspace |
| `--page-number` | `int64` | No | `1` | Page number (starts at `1`) |
| `--page-size` | `int64` | No | `20` | Page size (`1`–`100`) |

```sh
gumlet video asset list-recoverable --workspace-id WORKSPACE_ID -o table
```

---

### `gumlet video asset recover`

Recover a deleted video asset from the recycle bin.

- **API:** `POST /video/asset/recover`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--asset-id` | `string` | Yes | | ID of the deleted asset |

```sh
gumlet video asset recover --asset-id DELETED_ASSET_ID
```

---

## Video Playlists

### `gumlet video playlist list`

List all playlists in a workspace.

- **API:** `GET /video/playlist`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | Workspace ID to list playlists for |

```sh
gumlet video playlist list --workspace-id WORKSPACE_ID -o table
```

---

### `gumlet video playlist create`

Create a new playlist.

- **API:** `POST /video/playlist`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | Workspace (collection) ID for the playlist |
| `--title` | `string` | Yes | | Title of the playlist |
| `--description` | `string` | No | | Description of the playlist |

```sh
gumlet video playlist create --workspace-id WORKSPACE_ID --title "Getting Started"
```

---

### `gumlet video playlist update`

Update a playlist's title, description, or channel visibility.

- **API:** `POST /video/playlist/{playlist_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--playlist-id` | `string` | Yes | | ID of the playlist to update |
| `--title` | `string` | No | | New title for the playlist |
| `--description` | `string` | No | | New description for the playlist |
| `--channel-visibility` | `bool` | No | `false` | Show the playlist on the channel page (`true` or `false`) |

**Constraints:** At least one of `--title`, `--description`, or `--channel-visibility` is required.

```sh
gumlet video playlist update --playlist-id PLAYLIST_ID --channel-visibility=false
```

---

### `gumlet video playlist delete`

Delete a playlist.

- **API:** `DELETE /video/playlist/{playlist_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--playlist-id` | `string` | Yes | | ID of the playlist to delete |

```sh
gumlet video playlist delete --playlist-id PLAYLIST_ID
```

---

### `gumlet video playlist get-assets`

Get assets in a playlist.

- **API:** `GET /video/playlist/{playlist_id}/assets`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--playlist-id` | `string` | Yes | | ID of the playlist |

```sh
gumlet video playlist get-assets --playlist-id PLAYLIST_ID -o table
```

---

### `gumlet video playlist add-asset`

Add one or more assets to a playlist.

- **API:** `POST /video/playlist/{playlist_id}/asset`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--playlist-id` | `string` | Yes | | ID of the playlist |
| `--asset-ids` | `stringSlice` | Yes | | Comma-separated asset IDs to add |
| `--positions` | `intSlice` | No | | Comma-separated positions corresponding to each asset ID (count must match `--asset-ids`) |

```sh
gumlet video playlist add-asset --playlist-id PLAYLIST_ID --asset-ids ASSET_1,ASSET_2 --positions 0,1
```

---

### `gumlet video playlist remove-asset`

Remove one or more assets from a playlist.

- **API:** `DELETE /video/playlist/{playlist_id}/asset`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--playlist-id` | `string` | Yes | | ID of the playlist |
| `--asset-ids` | `stringSlice` | Yes | | Asset IDs to remove from the playlist |

```sh
gumlet video playlist remove-asset --playlist-id PLAYLIST_ID --asset-ids ASSET_1,ASSET_2
```

---

### `gumlet video playlist reorder`

Move a single asset to a specific position or sort the entire playlist.

- **API:** `POST /video/playlists/{playlist_id}/reorder`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--playlist-id` | `string` | Yes | | ID of the playlist |
| `--asset-id` | `string` | Conditional | | ID of the asset to move |
| `--page-number` | `int64` | Conditional | | Playlist page number (starts at `1`) |
| `--page-size` | `int64` | Conditional | | Playlist page size (`1`–`200`) |
| `--asset-position` | `int64` | Conditional | | Zero-based target position within the page (`0` to `page-size - 1`) |
| `--sort-by` | `string` | Conditional | | Sort field: `title` or `created_at` |
| `--sort-order` | `string` | Conditional | | Sort direction: `asc` or `desc` |

**Constraints:** Supply either **move mode** (`--asset-id`, `--page-number`, `--page-size`, `--asset-position` all together) or **sort mode** (`--sort-by` and `--sort-order` together). The two modes are mutually exclusive.

```sh
# Move a single asset
gumlet video playlist reorder --playlist-id PLAYLIST_ID --asset-id ASSET_ID --page-number 1 --page-size 20 --asset-position 0

# Sort the entire playlist
gumlet video playlist reorder --playlist-id PLAYLIST_ID --sort-by title --sort-order asc
```

---

## Channel Viewers

### `gumlet video channel viewer list`

List subscribers for a members-only channel.

- **API:** `GET /channel/{workspace_id}/viewers`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the video workspace |
| `--page-number` | `int64` | No | `1` | Page number (starts at `1`) |
| `--page-size` | `int64` | No | `20` | Page size (`1`–`100`) |
| `--next-page-token` | `string` | No | | Cursor returned by the previous page |

**Constraints:** `--page-number` and `--next-page-token` are mutually exclusive.

```sh
gumlet video channel viewer list --workspace-id WORKSPACE_ID --page-size 25 -o table
```

---

### `gumlet video channel viewer invite`

Invite viewers to a members-only channel.

- **API:** `POST /channel/{video_workspace_id}/viewers/invite`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the video workspace |
| `--viewers-file` | `string` | Yes | | JSON file containing an array of 1–200 `{"email":"...","name":"..."}` objects (`-` for stdin; max 1 MiB) |

**Validation:** Each entry must have a bare email address and a non-empty name of 1–100 characters.

```sh
gumlet video channel viewer invite --workspace-id WORKSPACE_ID --viewers-file ./viewers.json
```

---

### `gumlet video channel viewer remove`

Remove subscribers from a channel.

- **API:** `POST /channel/{video_workspace_id}/viewers/remove`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--workspace-id` | `string` | Yes | | ID of the video workspace |
| `--emails` | `stringSlice` | Yes | | 1–200 comma-separated bare email addresses |

```sh
gumlet video channel viewer remove --workspace-id WORKSPACE_ID --emails viewer@example.com
```

---

## Live Workspaces

> **Security note:** Live JSON responses can contain stream keys, ingest URLs, or presigned upload URLs. Keep them out of shared logs; `-o table` omits sensitive credentials.

### `gumlet video live workspace list`

List live streaming workspaces.

- **API:** `GET /video/sources/live`
- *No command-specific flags.*

```sh
gumlet video live workspace list -o table
```

---

### `gumlet video live workspace create`

Create a live streaming workspace.

- **API:** `POST /video/sources/live`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--name` | `string` | Yes | | Name of the live workspace |

```sh
gumlet video live workspace create --name "Studio"
```

---

### `gumlet video live workspace update`

Update a live streaming workspace.

- **API:** `POST /video/sources/live/{live_workspace_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-workspace-id` | `string` | Yes | | ID of the live workspace |
| `--name` | `string` | Conditional | | New name |
| `--video-workspace-id` | `string` | Conditional | | ID of the associated video-on-demand workspace |

**Constraints:** At least one of `--name` or `--video-workspace-id` is required.

```sh
gumlet video live workspace update --live-workspace-id LIVE_WORKSPACE_ID --name "Studio A"
```

---

### `gumlet video live workspace delete`

Delete a live streaming workspace.

- **API:** `DELETE /video/sources/live/{live_workspace_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-workspace-id` | `string` | Yes | | ID of the live workspace |

```sh
gumlet video live workspace delete --live-workspace-id LIVE_WORKSPACE_ID
```

---

## Live Assets

### `gumlet video live asset list`

List live streams in a live workspace.

- **API:** `GET /video/live/assets/list/{live_source_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-workspace-id` | `string` | Yes | | ID of the live workspace |
| `--status` | `stringSlice` | No | | Comma-separated statuses: `created`, `active`, `complete`, `disconnected`, `errored`, `deleted` |
| `--page-number` | `int64` | No | `1` | Page number (starts at `1`) |
| `--page-size` | `int64` | No | `20` | Page size (`1`–`100`) |
| `--next-page-token` | `string` | No | | Cursor returned by the previous page |

**Constraints:** `--page-number` and `--next-page-token` are mutually exclusive.

```sh
gumlet video live asset list --live-workspace-id LIVE_WORKSPACE_ID --status active,created -o table
```

---

### `gumlet video live asset get`

Get live stream status and details.

- **API:** `GET /video/live/assets/{live_asset_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |

```sh
gumlet video live asset get --live-asset-id LIVE_ASSET_ID -o table
```

---

### `gumlet video live asset create`

Create a live stream.

- **API:** `POST /video/live/assets`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-workspace-id` | `string` | Yes | | ID of the live workspace |
| `--resolution` | `string` | Yes | | HLS resolution: `240p`, `360p`, `480p`, `540p`, `720p`, or `1080p` |
| `--title` | `string` | No | | Stream title |
| `--mp4-access` | `bool` | No | `false` | Create a downloadable MP4 recording |
| `--orientation` | `string` | No | | Video orientation: `landscape` or `portrait` |
| `--start-at` | `string` | No | | Scheduled start time in RFC 3339 format (e.g. `2026-12-01T15:00:00Z`) |

```sh
gumlet video live asset create --live-workspace-id LIVE_WORKSPACE_ID --resolution 720p --title "Live Demo"
```

---

### `gumlet video live asset update`

Update a live stream's title or scheduled start time.

- **API:** `POST /video/live/assets/update`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |
| `--title` | `string` | Conditional | | New stream title |
| `--start-at` | `string` | Conditional | | New scheduled start time in RFC 3339 format |

**Constraints:** At least one of `--title` or `--start-at` is required.

```sh
gumlet video live asset update --live-asset-id LIVE_ASSET_ID --title "Updated Demo"
```

---

### `gumlet video live asset delete`

Delete a live stream.

- **API:** `DELETE /video/live/assets/{live_asset_id}`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |

```sh
gumlet video live asset delete --live-asset-id LIVE_ASSET_ID
```

---

### `gumlet video live asset start`

Start a live stream.

- **API:** `POST /video/live/assets/{live_asset_id}/start`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |

```sh
gumlet video live asset start --live-asset-id LIVE_ASSET_ID
```

---

### `gumlet video live asset complete`

Complete (end) a live stream.

- **API:** `POST /video/live/assets/{live_asset_id}/complete`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |

```sh
gumlet video live asset complete --live-asset-id LIVE_ASSET_ID
```

---

### `gumlet video live asset history`

Get live stream status history.

- **API:** `GET /video/live/assets/{live_asset_id}/history`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |

```sh
gumlet video live asset history --live-asset-id LIVE_ASSET_ID -o table
```

---

### `gumlet video live asset thumbnail-upload-urls`

Request presigned thumbnail upload URLs for a live stream. Returns upload URLs in JSON output (`-o json`); does not upload files.

- **API:** `POST /video/live/assets/thumbnail/upload`

| Flag | Type | Required | Default | Description |
| :--- | :--- | :--- | :--- | :--- |
| `--live-asset-id` | `string` | Yes | | ID of the live asset |
| `--statuses` | `stringSlice` | Yes | | Comma-separated thumbnail states: `preparing`, `disconnected`, `end` |

```sh
gumlet video live asset thumbnail-upload-urls --live-asset-id LIVE_ASSET_ID --statuses preparing,end
```
