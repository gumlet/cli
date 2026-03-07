---
name: gumlet-image-purge
description: Purge the Gumlet CDN cache for an image source subdomain using the Gumlet CLI. Use this skill when you need to invalidate cached images after updating source files — either purge the entire cache for a subdomain or target specific image URLs/paths. Requires the Gumlet CLI to be installed and authenticated via `gumlet login`.
---

## Purge all cached images for a subdomain

```sh
gumlet image purge --subdomain mycompany
```

## Purge specific image paths

```sh
gumlet image purge --subdomain mycompany \
  --urls "https://mycompany.gumlet.io/photo.jpg,https://mycompany.gumlet.io/banner.png"
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--subdomain` | ✅ | Subdomain to purge cache for (e.g. `mycompany` for `mycompany.gumlet.io`) |
| `--urls` | ✅ | Comma-separated list of specific URLs to purge |

## Notes

- Targeted purges (with `--urls`) are faster and only invalidate the listed images.
