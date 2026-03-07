---
name: gumlet-video-asset-upload
description: "Upload a local video file to Gumlet as a new video asset using the Gumlet CLI. Use this skill when you need to upload a video file from the local filesystem to Gumlet for transcoding and delivery. The CLI handles the two-step process automatically: creating the upload session and streaming the file to the returned pre-signed URL. A real-time progress indicator is shown during upload. Requires the Gumlet CLI to be installed and authenticated via gumlet login."
---

## Basic upload

Title defaults to the filename when `--title` is not provided.

```sh
gumlet video asset upload --file ./video.mp4 --workspace-id <id>
```

## Upload with metadata

```sh
gumlet video asset upload \
  --file ./video.mp4 \
  --workspace-id <id> \
  --title "Product Demo" \
  --description "Demo video for Q2 launch" \
  --format ABR \
  --tag demo,product
```

## Upload and add directly to a playlist

```sh
gumlet video asset upload \
  --file ./video.mp4 \
  --workspace-id <id> \
  --playlist-id <playlist_id>
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--file` | ✅ | Path to the local video file |
| `--workspace-id` | ✅ | Workspace (collection) ID |
| `--format` | ❌ | `ABR` (HLS + DASH) or `MP4`. Default: `ABR` |
| `--title` | ❌ | Asset title. Defaults to filename |
| `--description` | ❌ | Asset description |
| `--playlist-id` | ❌ | Add asset to this playlist after upload |
| `--tag` | ❌ | Comma-separated tags |

## After upload

Check processing status with:

```sh
gumlet video asset get --asset-id <returned_asset_id>
```
