---
name: gumlet-video-asset
description: Manage Gumlet video assets using the Gumlet CLI. Use this skill to list assets in a workspace with optional filters, get full details of a single asset including playback and thumbnail URLs, or delete an asset permanently. Requires the Gumlet CLI to be installed and authenticated via `gumlet login`.
---

## List assets in a workspace

```sh
gumlet video asset list --workspace-id <id>
gumlet video asset list --workspace-id <id> --output table
```

### Optional filters

| Flag | Description |
|---|---|
| `--status` | Filter by status: `ready`, `processing`, `upload-pending` |
| `--tag` | Filter by tag |
| `--title` | Filter by title |
| `--folder` | Filter by folder |
| `--size` | Page size |
| `--offset` | Pagination offset |
| `--sort-by` | Field to sort by |
| `--order-by` | `asc` or `desc` |
| `--playlist-id` | Filter by playlist |

```sh
gumlet video asset list --workspace-id <id> --status ready --size 20 --output table
gumlet video asset list --workspace-id <id> --tag promo
```

## Get details of a single asset

Returns asset ID, title, playback URL, thumbnail URL, status, tags, and collection info.

```sh
gumlet video asset get --asset-id <id>
```

## Delete an asset

```sh
gumlet video asset delete --asset-id <id>
```

## Notes

- `--workspace-id` is required for `list`.
- `--asset-id` is required for `get` and `delete`.
- Use `--output table` for a human-readable view.
