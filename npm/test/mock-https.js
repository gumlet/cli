// Replace only the installer transport. No test should contact GitHub.
const https = require("https");
const fs = require("fs");
const path = require("path");
const { EventEmitter } = require("events");
const { PassThrough, Readable } = require("stream");

function mockHTTPS(route) {
  const original = https.get;
  https.get = (url, options, callback) => {
    const request = new EventEmitter();
    let response;
    let destroyed = false;
    request.destroy = (error) => {
      destroyed = true;
      if (response) response.destroy();
      if (error) request.emit("error", error);
    };
    queueMicrotask(() => {
      if (destroyed) return;
      try {
        const { statusCode = 200, headers = {}, body, stall } = route(url);
        if (stall === "request") return;
        response = stall === "response" ? new PassThrough() : Readable.from([body ?? Buffer.alloc(0)]);
        Object.assign(response, { statusCode, headers });
        callback(response);
      } catch (error) {
        request.emit("error", error);
      }
    });
    return request;
  };
  return () => { https.get = original; };
}

function mockAssets(dir) {
  return mockHTTPS((url) => {
    const name = path.posix.basename(url.pathname);
    if (url.origin === "https://github.com" && /^\/gumlet\/cli\/releases\/download\/v[^/]+\/[^/]+$/.test(url.pathname)) {
      return { statusCode: 302, headers: { location: `https://fixtures.test/assets/${name}` } };
    }
    if (url.origin === "https://fixtures.test" && url.pathname === `/assets/${name}`) {
      return { statusCode: 307, headers: { location: `../files/${name}` } };
    }
    if (url.origin === "https://fixtures.test" && url.pathname === `/files/${name}`) {
      return { body: fs.readFileSync(path.join(dir, name)) };
    }
    throw new Error(`Unexpected fixture URL: ${url}`);
  });
}

if (process.env.GUMLET_TEST_ASSETS) mockAssets(process.env.GUMLET_TEST_ASSETS);

module.exports = { mockAssets, mockHTTPS };
