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
| `--folder-id` | Filter by folder ID; pass `--folder-id=` for unfiled assets |
| `--page-size` | Results per page, 1–100 |
| `--page-number` | Page number, starting at 1 |
| `--next-page-token` | Cursor from the last JSON response, instead of a page number |
| `--sort-by` | `created_at`, `uploaded_at`, `title`, `duration`, or `status` |
| `--order-by` | `asc` or `desc` |
| `--playlist-id` | Filter by playlist |

```sh
gumlet video asset list --workspace-id <id> --status ready --page-size 20 --output table
gumlet video asset list --workspace-id <id> --tag promo
```

## Get details of a single asset

Returns the full API response in JSON mode, including nested fields for the title and output URLs when available.

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
- Cursor pagination requires the default or `created_at` sort and no title filter. Repeat the filters and page size when passing a cursor.
