#!/usr/bin/env node
// Downloads the correct gumlet binary from GitHub Releases on postinstall.

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");
const { version } = require("./package.json");

const REPO = "gumlet/cli";
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, process.platform === "win32" ? "gumlet.exe" : "gumlet");

function platformArchive() {
  const platform = process.platform;
  const arch = process.arch;

  const os = platform === "darwin" ? "darwin" : platform === "win32" ? "windows" : "linux";
  const goArch = arch === "arm64" ? "arm64" : "x86_64";
  const ext = platform === "win32" ? "zip" : "tar.gz";

  return `gumlet_${os}_${goArch}.${ext}`;
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (u) => {
      https.get(u, { headers: { "User-Agent": "gumlet-cli-installer" } }, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return follow(res.headers.location);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${u}`));
        }
        const file = fs.createWriteStream(dest);
        res.pipe(file);
        file.on("finish", () => file.close(resolve));
        file.on("error", reject);
      }).on("error", reject);
    };
    follow(url);
  });
}

async function main() {
  const archive = platformArchive();
  const url = `https://github.com/${REPO}/releases/download/v${version}/${archive}`;
  const tmpArchive = path.join(BIN_DIR, archive);

  fs.mkdirSync(BIN_DIR, { recursive: true });

  console.log(`Downloading gumlet v${version} from GitHub Releases...`);
  await downloadFile(url, tmpArchive);

  console.log("Extracting binary...");
  if (archive.endsWith(".zip")) {
    execSync(`unzip -o "${tmpArchive}" gumlet.exe -d "${BIN_DIR}"`);
  } else {
    execSync(`tar -xzf "${tmpArchive}" -C "${BIN_DIR}" gumlet`);
  }

  fs.unlinkSync(tmpArchive);
  fs.chmodSync(BIN_PATH, 0o755);
  console.log("gumlet installed successfully.");
}

main().catch((err) => {
  console.error("Failed to install gumlet:", err.message);
  process.exit(1);
});
