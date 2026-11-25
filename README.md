# Gumlet CLI

Manage [Gumlet](https://www.gumlet.com) video workspaces, assets, playlists, live streams, channel viewers, image sources, and cache purges from your terminal.

## Install

**Homebrew** (macOS):

```sh
brew install --cask gumlet/tap/gumlet
```

**npm** (macOS, Linux, Windows):

```sh
npm install -g @gumlet/cli
```

**Manual download:** Download the archive for your OS/architecture (`x86_64` or `arm64`) from [GitHub Releases](https://github.com/gumlet/cli/releases), verify against `checksums.txt`, and place `gumlet` on your `PATH`.

Verify with `gumlet --version`.

## Authenticate

Create a key in [Gumlet API Keys](https://dash.gumlet.com/developer/api-keys) and save it locally:

```sh
gumlet login
```

The key is stored in `~/.gumlet.yaml`. For CI or non-interactive environments, set `GUMLET_API_KEY` instead (takes precedence over the saved key). Run `gumlet logout` to remove the saved key.

## Quickstart

```sh
# List workspaces and image sources
gumlet video workspace list -o table
gumlet image source list -o table

# Upload a video and check its status
gumlet video asset upload --workspace-id WORKSPACE_ID --file ./intro.mp4 --title "Introduction"
gumlet video asset get --asset-id ASSET_ID -o table

# Purge cached images
gumlet image purge --source-id SOURCE_ID --urls "/photo.jpg,/banner.png"
```

See **[COMMANDS.md](COMMANDS.md)** for the full command reference (all subcommands, flags, arguments, validation rules, and examples), or run `gumlet <command> --help`.

## Output & Scripting

- API commands default to `--output json` (`-o json`) with raw API fields and pagination cursors.
- Pass `-o table` for concise human-readable tables, and `--no-color` to disable styling.
- Progress and errors go to stderr; command errors exit nonzero.

## AI Agent Skills

Agent skills for Gumlet CLI workflows live in [skills](./skills) and on [skills.sh](https://skills.sh/gumlet/cli):

```sh
npx skills add https://github.com/gumlet/cli --skill gumlet-video-asset
```

## Development

```sh
go run . --help
go test ./...
go vet ./...
```

Integration tests run against local mock servers and do not require an API key. For npm installer tests (Node.js 26):

```sh
npm --prefix npm ci --ignore-scripts
npm --prefix npm test
```

## License

[Apache 2.0](LICENSE) © [Gumlet Pte. Ltd.](https://www.gumlet.com)
