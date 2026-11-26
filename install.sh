#!/bin/sh
set -eu

REPO="gumlet/cli"
BINARY="gumlet"
VERSION="${GUMLET_VERSION:-}"
INSTALL_DIR="${GUMLET_INSTALL_DIR:-}"

fail() {
  printf 'Error: %s\n' "$1" >&2
  exit 1
}

detect_os() {
  case "$(uname -s)" in
    Linux)  echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)      fail "Unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64)  echo "x86_64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *)               fail "Unsupported architecture: $(uname -m)" ;;
  esac
}

choose_install_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    echo "$INSTALL_DIR"
    return
  fi
  if [ "$(id -u)" -eq 0 ] || [ -w "/usr/local/bin" ]; then
    echo "/usr/local/bin"
  else
    echo "${HOME}/.local/bin"
  fi
}

download() {
  url="$1"
  dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --proto '=https' --tlsv1.2 -o "$dest" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q --https-only -O "$dest" "$url"
  else
    fail "Neither curl nor wget is installed"
  fi
}

sha256_file() {
  file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print tolower($1)}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print tolower($1)}'
  elif command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$file" | awk '{print tolower($NF)}'
  else
    fail "No SHA-256 tool found (need sha256sum, shasum, or openssl)"
  fi
}

verify_checksum() {
  archive_path="$1"
  checksums_path="$2"
  archive_name="$3"

  expected="$(
    tr -d '\r' < "$checksums_path" | awk -v target="$archive_name" '
      {
        name = $2
        sub(/^\*/, "", name)
        if (name == target) {
          count++
          hash = tolower($1)
        }
      }
      END {
        if (count == 1 && hash ~ /^[0-9a-f]{64}$/) print hash
      }
    '
  )"

  [ -n "$expected" ] || fail "Could not find a valid SHA-256 checksum for ${archive_name}"

  actual="$(sha256_file "$archive_path")"
  if [ "$actual" != "$expected" ]; then
    fail "SHA-256 checksum mismatch for ${archive_name} (expected ${expected}, got ${actual})"
  fi
}

main() {
  os="$(detect_os)"
  arch="$(detect_arch)"
  install_dir="$(choose_install_dir)"
  archive="${BINARY}_${os}_${arch}.tar.gz"

  if [ -n "$VERSION" ]; then
    version="${VERSION#v}"
    case "$version" in
      *[!0-9A-Za-z._-]* | "")
        fail "Invalid GUMLET_VERSION: ${VERSION}"
        ;;
    esac
    base_url="https://github.com/${REPO}/releases/download/v${version}"
    label="v${version}"
  else
    base_url="https://github.com/${REPO}/releases/latest/download"
    label="latest"
  fi

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT INT TERM

  checksums_path="${tmpdir}/checksums.txt"
  archive_path="${tmpdir}/${archive}"

  printf 'Downloading %s (%s) for %s/%s...\n' "$BINARY" "$label" "$os" "$arch"
  download "${base_url}/checksums.txt" "$checksums_path"
  download "${base_url}/${archive}" "$archive_path"

  verify_checksum "$archive_path" "$checksums_path" "$archive"

  tar -xzf "$archive_path" -C "$tmpdir" "$BINARY"
  extracted="${tmpdir}/${BINARY}"
  if [ ! -f "$extracted" ] || [ -L "$extracted" ]; then
    fail "Archive did not contain a valid ${BINARY} executable"
  fi

  mkdir -p "$install_dir"
  staged="${install_dir}/.${BINARY}.tmp.$$"
  cp "$extracted" "$staged"
  chmod 0755 "$staged"
  mv -f "$staged" "${install_dir}/${BINARY}"

  printf 'Installed %s to %s/%s\n' "$BINARY" "$install_dir" "$BINARY"

  case ":${PATH}:" in
    *":${install_dir}:"*) ;;
    *)
      printf '\nNote: %s is not in your PATH. Add it with:\n' "$install_dir"
      printf '  export PATH="%s:$PATH"\n' "$install_dir"
      ;;
  esac
}

main "$@"
