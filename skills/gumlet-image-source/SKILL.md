---
name: gumlet-image-source
description: Manage Gumlet image sources using the Gumlet CLI. Use this skill to list all image sources, get details of a specific source, create a new image source connected to a storage backend (Amazon S3, GCS, web proxy, DigitalOcean Spaces, Wasabi, Azure, Cloudflare, Backblaze, etc.), update an existing source's configuration, or delete a source. Requires the Gumlet CLI to be installed and authenticated via `gumlet login`.
---

## List all image sources

```sh
gumlet image source list
gumlet image source list --output table
```

## Get details of a source

```sh
gumlet image source get --source-id <id>
```

## Create a new image source

### Amazon S3

```sh
gumlet image source add \
  --namespace mycompany \
  --type amazon \
  --config '{"bucket_name":"my-bucket","bucket_region":"us-east-1","access_key":"KEY","secret":"SECRET"}'
```

### Web proxy

```sh
gumlet image source add \
  --namespace mycompany \
  --type proxy \
  --config '{"whitelisted_domains":["example.com"]}'
```

### Google Cloud Storage

```sh
gumlet image source add \
  --namespace mycompany \
  --type gcs \
  --config '{"bucket_name":"my-gcs-bucket","service_account_json":"{...}"}'
```

### DigitalOcean Spaces

```sh
gumlet image source add \
  --namespace mycompany \
  --type dostorage \
  --config '{"bucket_name":"my-space","bucket_region":"nyc3","access_key":"KEY","secret":"SECRET"}'
```

## Update an existing source

```sh
gumlet image source update --source-id <id> --type amazon --config '{"bucket_name":"new-bucket"}'
```

## Delete a source

```sh
gumlet image source delete --source-id <id>
```

## Flags

| Flag | Required | Description |
|---|---|---|
| `--namespace` | ✅ (add only) | Subdomain (e.g. `mycompany` → `mycompany.gumlet.io`) |
| `--type` | ✅ (add) | Source type: `amazon`, `proxy`, `gcs`, `dostorage`, `wasabi`, `cloudinary`, `azure`, `linode`, `backblaze`, `cloudflare` |
| `--config` | ❌ | JSON string with type-specific storage configuration |
| `--source-id` | ✅ (get/update/delete) | ID of the source |

## Notes

- Use `--output table` for a human-readable view.
- The `--config` JSON structure varies by source type.
