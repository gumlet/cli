# Gumlet CLI

Manage [Gumlet](https://www.gumlet.com) from your terminal. Upload videos,
organize assets and playlists, manage live streams, and purge cached images.
Commands return JSON by default for scripting; use `-o table` for readable output.

## Install

**npm — macOS, Linux, and Windows**

```sh
npm install -g @gumlet/cli
```

Requires Node.js 14.14 or later. The installer downloads and verifies the
matching binary from GitHub Releases. macOS and Linux support x86-64 and
ARM64; Windows supports x86-64.

**Homebrew — macOS**

```sh
brew install --cask gumlet/tap/gumlet
```

**Manual download — no Node.js required**

Download the archive for your OS and architecture from
[GitHub Releases](https://github.com/gumlet/cli/releases). Verify its SHA-256
digest against `checksums.txt`, extract it, and put `gumlet` (or `gumlet.exe`
on Windows) on your `PATH`.

## Get started

Sign in with your Gumlet account via browser:

```sh
gumlet --version
gumlet login
gumlet user get -o table
gumlet --help
```

For CI or non-interactive environments, create an API key in
[Gumlet API Keys](https://dash.gumlet.com/developer/api-keys) and set the
`GUMLET_API_KEY` environment variable instead.

For more examples and configuration details, read the
[full README in the Gumlet CLI GitHub repository](https://github.com/gumlet/cli#readme).
