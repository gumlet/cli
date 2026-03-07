---
name: gumlet-video-playlist
description: Manage Gumlet video playlists using the Gumlet CLI. Use this skill to list playlists in a workspace, create or update a playlist, add or remove assets from a playlist, view assets inside a playlist, or delete a playlist. Requires the Gumlet CLI to be installed and authenticated via `gumlet login`.
---

## List playlists in a workspace

```sh
gumlet video playlist list --workspace-id <id>
gumlet video playlist list --workspace-id <id> --output table
```

## Create a playlist

```sh
gumlet video playlist create --workspace-id <id> --title "Best of 2026"
gumlet video playlist create --workspace-id <id> --title "Best of 2026" --description "Top videos of the year"
```

## Update a playlist

All flags are optional — only supplied flags are updated.

```sh
gumlet video playlist update --playlist-id <id> --title "Renamed Playlist"
gumlet video playlist update --playlist-id <id> --description "Updated description"
gumlet video playlist update --playlist-id <id> --channel-visibility public
```

## Get assets in a playlist

```sh
gumlet video playlist get-assets --playlist-id <id>
gumlet video playlist get-assets --playlist-id <id> --output table
```

## Add assets to a playlist

```sh
# Add a single asset
gumlet video playlist add-asset --playlist-id <id> --asset-ids asset_abc

# Add multiple assets
gumlet video playlist add-asset --playlist-id <id> --asset-ids asset_1,asset_2,asset_3

# Add multiple assets with explicit positions
gumlet video playlist add-asset --playlist-id <id> --asset-ids asset_1,asset_2 --positions 1,2
```

> When `--positions` is provided, the count must match the number of `--asset-ids`.

## Remove assets from a playlist

```sh
gumlet video playlist remove-asset --playlist-id <id> --asset-ids asset_1,asset_2
```

## Delete a playlist

```sh
gumlet video playlist delete --playlist-id <id>
```

## Notes

- `--workspace-id` is required for `list` and `create`.
- `--playlist-id` is required for all other subcommands.
- Use `--output table` for a human-readable view.
