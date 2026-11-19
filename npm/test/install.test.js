const assert = require("node:assert/strict");
const { execFile } = require("node:child_process");
const { createHash } = require("node:crypto");
const fs = require("node:fs/promises");
const os = require("node:os");
const path = require("node:path");
const { after, before, test } = require("node:test");
const { promisify } = require("node:util");
const { downloadFile, install, platformArchive } = require("../install");
const { mockAssets, mockHTTPS } = require("./mock-https");

const exec = promisify(execFile);
const pkg = path.resolve(__dirname, "..");
const win = process.platform === "win32";
const fixtureExe = win ? "fixture.exe" : "fixture";
const versions = [
  ["darwin", "arm64", "gumlet_darwin_arm64.tar.gz"],
  ["darwin", "x64", "gumlet_darwin_x86_64.tar.gz"],
  ["linux", "arm64", "gumlet_linux_arm64.tar.gz"],
  ["linux", "x64", "gumlet_linux_x86_64.tar.gz"],
  ["win32", "x64", "gumlet_windows_x86_64.zip"],
];
let workspace;
let fixture;

before(async () => {
  workspace = await fs.mkdtemp(path.join(os.tmpdir(), "gumlet npm tests-"));
  fixture = path.join(workspace, fixtureExe);
  await exec("go", ["build", "-o", fixture, path.join(__dirname, "fixture.go")]);
});

after(async () => {
  if (workspace) await fs.rm(workspace, { recursive: true, force: true });
});

async function assets(parent, name, variant, badChecksum = false) {
  const dir = path.join(parent, "assets");
  await fs.mkdir(dir);
  const archive = path.join(dir, name);
  if (badChecksum || variant === "corrupt") await fs.writeFile(archive, "not even an archive");
  else await exec(fixture, ["archive", archive, variant]);
  const digest = badChecksum ? "0".repeat(64) : createHash("sha256").update(await fs.readFile(archive)).digest("hex");
  await fs.writeFile(path.join(dir, "checksums.txt"), `${"f".repeat(64)}  other-file.zip\r\n${digest}  ${name}\r\n`);
  return dir;
}

test("only selects platforms built by GoReleaser", async () => {
  for (const [platform, arch, name] of versions) assert.equal(platformArchive(platform, arch), name);
  const restore = mockHTTPS(() => assert.fail("unsupported platform made a request"));
  try {
    for (const [platform, arch] of [["win32", "arm64"], ["linux", "ia32"], ["freebsd", "x64"]]) {
      await assert.rejects(install({ platform, arch, binDir: path.join(workspace, "bin") }),
        { message: `Unsupported platform: ${platform}/${arch}` });
    }
  } finally {
    restore();
  }
  assert.deepEqual(await fs.readdir(workspace), [fixtureExe]);
});

for (const [platform, arch] of [["linux", "x64"], ["win32", "x64"]]) {
  const binary = platform === "win32" ? "gumlet.exe" : "gumlet";
  const archive = platformArchive(platform, arch);
  test(`${archive}: extracts only the verified executable`, async () => {
    const root = await fs.mkdtemp(path.join(workspace, "extract-"));
    const binDir = path.join(root, "bin");
    const restore = mockAssets(await assets(root, archive, "success"));
    try {
      await install({ platform, arch, binDir });
      // Reinstallation must also replace an existing binary.
      await fs.writeFile(path.join(binDir, binary), "old binary");
      await install({ platform, arch, binDir });
    } finally {
      restore();
    }
    assert.deepEqual((await fs.readdir(root)).sort(), ["assets", "bin"]);
    assert.deepEqual(await fs.readdir(binDir), [binary]);
    assert.deepEqual(await fs.readFile(path.join(binDir, binary)), await fs.readFile(fixture));
    if (!win) assert.equal((await fs.stat(path.join(binDir, binary))).mode & 0o777, 0o755);
  });

  for (const [variant, message] of [
    ["bad-checksum", /SHA-256 checksum mismatch/],
    ["corrupt", platform === "win32" ? /central directory/i : /incorrect header/i],
    ["duplicate", /Invalid or duplicate/],
    ["symlink", /Invalid or duplicate/],
    ["missing", /Archive does not contain/],
  ]) {
    test(`${archive}: rejects ${variant} and cleans staging`, async () => {
      const root = await fs.mkdtemp(path.join(workspace, "failure-"));
      const restore = mockAssets(await assets(root, archive, variant, variant === "bad-checksum"));
      try {
        await assert.rejects(install({ platform, arch, binDir: path.join(root, "bin") }), message);
      } finally {
        restore();
      }
      assert.deepEqual(await fs.readdir(root), ["assets"]);
    });
  }
}

test("downloads enforce HTTPS, redirects, timeouts and size bounds", async () => {
  const cases = [
    [() => ({ statusCode: 301, headers: { location: "http://fixtures.test/insecure" } }), /require HTTPS/],
    [() => ({ statusCode: 302, headers: { location: "/loop" } }), /Too many.*redirects/],
    [() => ({ statusCode: 307 }), /missing Location/],
    [() => ({ statusCode: 404 }), /HTTP 404/],
    [() => ({ stall: "request" }), /Download timed out/],
    [() => ({ stall: "response" }), /Download timed out/],
    [() => ({ body: Buffer.alloc(100) }), /Download exceeds/],
  ];
  for (const [index, [route, message]] of cases.entries()) {
    const restore = mockHTTPS(route);
    try {
      await assert.rejects(downloadFile("https://fixtures.test/start", path.join(workspace, `response-${index}`), 10, 100), message);
    } finally {
      restore();
    }
  }
});

test("failed downloads clean staging without touching an existing binary", async () => {
  const root = await fs.mkdtemp(path.join(workspace, "download-failure-"));
  const binDir = path.join(root, "bin");
  const binary = path.join(binDir, "gumlet");
  await fs.mkdir(binDir);
  await fs.writeFile(binary, "previous binary");
  const restore = mockHTTPS((url) => url.pathname.endsWith("/checksums.txt")
    ? { body: Buffer.from("fixture checksum") }
    : { statusCode: 503 });
  try {
    await assert.rejects(install({ platform: "linux", arch: "x64", binDir }), /HTTP 503/);
  } finally {
    restore();
  }
  assert.deepEqual(await fs.readdir(root), ["bin"]);
  assert.equal(await fs.readFile(binary, "utf8"), "previous binary");
});

test("packs, installs and invokes the npm shim offline", async () => {
  const root = await fs.mkdtemp(path.join(workspace, "package-"));
  const releaseAssets = await assets(root, platformArchive(), "success");
  const npm = (args, options) => exec(process.execPath, [process.env.npm_execpath, ...args], { cwd: pkg, ...options });
  const { stdout } = await npm(["pack", "--offline", "--ignore-scripts", "--json", "--pack-destination", root]);
  const [packed] = JSON.parse(stdout);
  assert.deepEqual(packed.files.map(({ path }) => path).sort(), ["gumlet", "install.js", "npm-shrinkwrap.json", "package.json"]);

  // A consumer lock reuses the dependency tarballs cached by npm ci, with
  // no fetching of registry metadata or release assets during the smoke test.
  const lock = require("../npm-shrinkwrap.json");
  const { "": published, ...dependencies } = lock.packages;
  const project = { name: "gumlet-npm-smoke", dependencies: { [published.name]: `file:${packed.filename}` } };
  const consumerLock = {
    name: project.name,
    lockfileVersion: 3,
    packages: {
      "": project,
      ...dependencies,
      [`node_modules/${published.name}`]: {
        ...published,
        resolved: project.dependencies[published.name],
        integrity: packed.integrity,
      },
    },
  };
  await fs.writeFile(path.join(root, "package.json"), JSON.stringify(project));
  await fs.writeFile(path.join(root, "package-lock.json"), JSON.stringify(consumerLock));
  const preload = path.join(__dirname, "mock-https.js").replace(/\\/g, "/");
  await npm(["ci", "--offline", "--no-audit", "--no-fund", "--ignore-scripts=false"], {
    cwd: root,
    env: { ...process.env, NODE_OPTIONS: `--require="${preload}"`, GUMLET_TEST_ASSETS: releaseAssets },
  });

  const shimDir = path.join(root, "node_modules", ".bin");
  const command = win ? "cmd.exe" : path.join(shimDir, "gumlet");
  const args = win ? ["/d", "/c", "gumlet.cmd", "fixture", "smoke"] : ["fixture", "smoke"];
  await assert.rejects(exec(command, args, { cwd: shimDir }), (err) => {
    assert.equal(err.code, 23, err.stderr);
    assert.equal(err.stdout.trim(), "fixture|smoke");
    return true;
  });
});
