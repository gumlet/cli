#!/usr/bin/env node
// Downloads the correct gumlet binary from GitHub Releases on postinstall.

const https = require("https");
const fs = require("fs");
const path = require("path");
const { createHash } = require("crypto");
const { once } = require("events");
const { Transform, pipeline: streamPipeline } = require("stream");
const { promisify } = require("util");
const { createGunzip } = require("zlib");
const tar = require("tar-stream");
const yauzl = require("yauzl");
const { version } = require("./package.json");

const pipeline = promisify(streamPipeline);
const CHECKSUM_LIMIT = 64 * 1024;
const ARCHIVE_LIMIT = 128 * 1024 * 1024;
const DOWNLOAD_TIMEOUT = 60_000;
const MAX_REDIRECTS = 5;
const REDIRECTS = new Set([301, 302, 303, 307, 308]);

// Keep these pairs and filenames in sync with .goreleaser.yaml.
const ARCHIVES = new Map([
  ["darwin/x64", "gumlet_darwin_x86_64.tar.gz"],
  ["darwin/arm64", "gumlet_darwin_arm64.tar.gz"],
  ["linux/x64", "gumlet_linux_x86_64.tar.gz"],
  ["linux/arm64", "gumlet_linux_arm64.tar.gz"],
  ["win32/x64", "gumlet_windows_x86_64.zip"],
  ["win32/arm64", "gumlet_windows_arm64.zip"],
]);

function platformArchive(platform = process.platform, arch = process.arch) {
  const archive = ARCHIVES.get(`${platform}/${arch}`);
  if (!archive) throw new Error(`Unsupported platform: ${platform}/${arch}`);
  return archive;
}

function byteLimit(maxBytes, label) {
  let received = 0;
  return new Transform({
    transform(chunk, encoding, callback) {
      received += chunk.length;
      if (received > maxBytes) return callback(new Error(`${label} exceeds ${maxBytes} bytes`));
      callback(null, chunk);
    },
  });
}

async function downloadFile(url, dest, maxBytes, timeoutMs = DOWNLOAD_TIMEOUT) {
  let current = new URL(url);
  let request;
  let response;
  const timer = setTimeout(() => {
    const error = new Error(`Download timed out after ${timeoutMs} ms`);
    if (response) response.destroy(error);
    if (request) request.destroy(error);
  }, timeoutMs);
  try {
    for (let redirects = 0; ; redirects++) {
      if (current.protocol !== "https:" || current.username || current.password) {
        throw new Error("Release downloads require HTTPS URLs without credentials");
      }
      response = await new Promise((resolve, reject) => {
        request = https.get(current, { headers: { "User-Agent": "gumlet-cli-installer" } }, resolve);
        request.on("error", reject);
      });
      if (REDIRECTS.has(response.statusCode)) {
        const location = response.headers.location;
        response.destroy();
        response = undefined;
        if (!location) throw new Error("Release download redirect is missing Location");
        if (redirects === MAX_REDIRECTS) throw new Error("Too many release download redirects");
        current = new URL(location, current);
        continue;
      }
      if (response.statusCode !== 200) {
        response.destroy();
        throw new Error(`HTTP ${response.statusCode} downloading ${new URL(url).pathname}`);
      }
      await pipeline(response, byteLimit(maxBytes, "Download"), fs.createWriteStream(dest, { flags: "wx" }));
      return;
    }
  } finally {
    clearTimeout(timer);
  }
}

async function verifyArchive(archivePath, checksumPath, archiveName) {
  const checksums = await fs.promises.readFile(checksumPath, "utf8");
  const matches = checksums.split(/\r?\n/).map((line) => /^([a-f0-9]{64})[ \t]+\*?(.+)$/i.exec(line))
    .filter((match) => match && match[2] === archiveName);
  if (matches.length !== 1) throw new Error(`Expected one SHA-256 checksum for ${archiveName}`);

  const hash = createHash("sha256");
  for await (const chunk of fs.createReadStream(archivePath)) hash.update(chunk);
  if (hash.digest("hex") !== matches[0][1].toLowerCase()) {
    throw new Error(`SHA-256 checksum mismatch for ${archiveName}`);
  }
}

async function extractTar(archivePath, executable, dest) {
  const extract = tar.extract();
  let found = false;
  const entries = (async () => {
    for await (const entry of extract) {
      if (entry.header.name !== executable) {
        for await (const chunk of entry) { /* drain ignored entries */ }
        continue;
      }
      if (found || entry.header.type !== "file" || entry.header.size > ARCHIVE_LIMIT) {
        throw new Error(`Invalid or duplicate ${executable} in archive`);
      }
      await pipeline(entry, fs.createWriteStream(dest, { flags: "wx" }));
      found = true;
    }
  })();
  // Wait for both sides to close on failure before deleting the staging directory.
  const results = await Promise.allSettled([
    entries,
    pipeline(fs.createReadStream(archivePath), createGunzip(), byteLimit(2 * ARCHIVE_LIMIT, "Expanded tar"), extract),
  ]);
  for (const result of results) {
    if (result.status === "rejected") throw result.reason;
  }
  if (!found) throw new Error(`Archive does not contain ${executable}`);
}

async function extractZip(archivePath, executable, dest) {
  const zip = await yauzl.openPromise(archivePath, { autoClose: false, strictFileNames: true });
  let found = false;
  try {
    for await (const entry of zip.eachEntry()) {
      if (entry.fileName !== executable) continue;
      const unixType = (entry.externalFileAttributes >>> 16) & 0o170000;
      const dosDirectory = (entry.externalFileAttributes & 0x10) !== 0;
      if (found || dosDirectory || (unixType !== 0 && unixType !== 0o100000) ||
          entry.uncompressedSize > ARCHIVE_LIMIT) {
        throw new Error(`Invalid or duplicate ${executable} in archive`);
      }
      await pipeline(await zip.openReadStreamPromise(entry), fs.createWriteStream(dest, { flags: "wx" }));
      found = true;
    }
  } finally {
    const closed = once(zip, "close");
    zip.close();
    await closed;
  }
  if (!found) throw new Error(`Archive does not contain ${executable}`);
}

async function install({
  platform = process.platform,
  arch = process.arch,
  binDir = path.join(__dirname, "bin"),
} = {}) {
  const archive = platformArchive(platform, arch);
  const executable = platform === "win32" ? "gumlet.exe" : "gumlet";
  const release = `https://github.com/gumlet/cli/releases/download/v${version}/`;
  const tempDir = await fs.promises.mkdtemp(path.join(path.dirname(binDir), ".gumlet-install-"));
  try {
    const checksumPath = path.join(tempDir, "checksums.txt");
    const archivePath = path.join(tempDir, archive);
    const stagedBinary = path.join(tempDir, executable);
    await downloadFile(`${release}checksums.txt`, checksumPath, CHECKSUM_LIMIT);
    await downloadFile(`${release}${archive}`, archivePath, ARCHIVE_LIMIT);
    await verifyArchive(archivePath, checksumPath, archive);
    const extract = archive.endsWith(".zip") ? extractZip : extractTar;
    await extract(archivePath, executable, stagedBinary);
    await fs.promises.chmod(stagedBinary, 0o755);
    await fs.promises.mkdir(binDir, { recursive: true });
    await fs.promises.rename(stagedBinary, path.join(binDir, executable));
  } finally {
    await fs.promises.rm(tempDir, { recursive: true, force: true, maxRetries: 3 });
  }
}

module.exports = { downloadFile, install, platformArchive };

if (require.main === module) {
  console.log(`Downloading and verifying gumlet v${version} from GitHub Releases...`);
  install().then(() => console.log("gumlet installed successfully.")).catch((err) => {
    console.error("Failed to install gumlet:", err.message);
    process.exitCode = 1;
  });
}
