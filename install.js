#!/usr/bin/env node

"use strict";

// notion-cli npm postinstall script.
//
// Security model:
//   1. Fetch SHA256SUMS for the release before any binary.
//   2. Download the binary, hashing as it streams.
//   3. Compare the computed hash to the entry in SHA256SUMS.
//   4. Only persist + chmod +x the binary if the hash matches; otherwise
//      delete any partial file and exit non-zero.
//
// Two failure classes, treated differently:
//   - Transient (network unreachable, 5xx, 404 because release hasn't
//     propagated, etc.): warn and exit 0 so `npm install` is not broken on
//     CI. The user can `npm run rebuild` later or build from source.
//   - Active-attack signal (downloaded bytes do not match the signed
//     checksum): hard-fail with non-zero exit. Never write the bad binary.
//
// All redirects are restricted to github.com and *.githubusercontent.com,
// and only https:// is followed. http:// redirects abort. Response bodies
// are capped at MAX_DOWNLOAD_BYTES.

const https = require("https");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const { URL } = require("url");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  arm64: "arm64",
  x64: "amd64",
};

const MAX_REDIRECTS = 5;
const MAX_DOWNLOAD_BYTES = 50 * 1024 * 1024; // 50 MB — far above any legit binary
const MAX_CHECKSUM_BYTES = 16 * 1024; // SHA256SUMS for 5 binaries is < 1 KB

function isAllowedHost(host) {
  if (!host) return false;
  const lower = host.toLowerCase();
  if (lower === "github.com") return true;
  if (lower === "objects.githubusercontent.com") return true;
  if (lower.endsWith(".githubusercontent.com")) return true;
  return false;
}

function getVersion() {
  try {
    const pkg = JSON.parse(
      fs.readFileSync(path.join(__dirname, "package.json"), "utf8")
    );
    return pkg.version;
  } catch {
    return "6.0.0";
  }
}

function checkPlatformPackage() {
  const platformKey = `${process.platform}-${process.arch}`;
  const pkgName = `@coastal-programs/notion-cli-${platformKey}`;
  try {
    require.resolve(`${pkgName}/package.json`);
    // Platform package exists, no need to download
    return true;
  } catch {
    return false;
  }
}

// fetchURL streams an https URL, enforcing redirect-host pinning, https-only,
// and a maximum-body size. The body is delivered to onData(chunk) and the
// returned promise resolves with { ok, status } once the stream ends.
function fetchURL(targetURL, onData) {
  return new Promise((resolve, reject) => {
    const visit = (urlString, redirects) => {
      let parsed;
      try {
        parsed = new URL(urlString);
      } catch (err) {
        reject(new Error(`invalid URL: ${err.message}`));
        return;
      }

      if (parsed.protocol !== "https:") {
        reject(
          new Error(
            `refusing to follow non-https URL (got ${parsed.protocol}//${parsed.host})`
          )
        );
        return;
      }
      if (!isAllowedHost(parsed.hostname)) {
        reject(new Error(`refusing redirect to disallowed host: ${parsed.hostname}`));
        return;
      }

      const req = https.get(
        urlString,
        { headers: { "User-Agent": "notion-cli-npm" } },
        (res) => {
          const status = res.statusCode || 0;
          if (status >= 300 && status < 400 && res.headers.location) {
            if (redirects >= MAX_REDIRECTS) {
              res.resume();
              reject(new Error("too many redirects"));
              return;
            }
            // Resolve relative redirects against the current URL.
            let next;
            try {
              next = new URL(res.headers.location, urlString).toString();
            } catch (err) {
              res.resume();
              reject(new Error(`invalid redirect target: ${err.message}`));
              return;
            }
            res.resume();
            visit(next, redirects + 1);
            return;
          }

          if (status !== 200) {
            res.resume();
            resolve({ ok: false, status });
            return;
          }

          let received = 0;
          let aborted = false;
          res.on("data", (chunk) => {
            if (aborted) return;
            received += chunk.length;
            if (received > MAX_DOWNLOAD_BYTES) {
              aborted = true;
              res.destroy();
              reject(
                new Error(
                  `response exceeded max size of ${MAX_DOWNLOAD_BYTES} bytes`
                )
              );
              return;
            }
            try {
              onData(chunk);
            } catch (err) {
              aborted = true;
              res.destroy();
              reject(err);
            }
          });
          res.on("end", () => {
            if (!aborted) {
              resolve({ ok: true, status });
            }
          });
          res.on("error", (err) => {
            if (!aborted) reject(err);
          });
        }
      );
      req.on("error", reject);
    };

    visit(targetURL, 0);
  });
}

async function fetchText(url, cap) {
  const limit = cap || MAX_CHECKSUM_BYTES;
  const chunks = [];
  let total = 0;
  const result = await fetchURL(url, (chunk) => {
    total += chunk.length;
    if (total > limit) {
      throw new Error(`text response exceeded ${limit} bytes`);
    }
    chunks.push(chunk);
  });
  if (!result.ok) {
    return { ok: false, status: result.status, body: "" };
  }
  return { ok: true, status: result.status, body: Buffer.concat(chunks).toString("utf8") };
}

// parseChecksums parses lines of "<hex>  <filename>" (sha256sum -b or text mode)
// and returns a map of filename -> hex.
function parseChecksums(text) {
  const out = {};
  for (const rawLine of text.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;
    // Format: "<64-hex-chars>  <filename>" (two spaces; '*' instead of ' ' for binary mode)
    const m = line.match(/^([0-9a-fA-F]{64})\s+\*?(.+)$/);
    if (!m) continue;
    out[m[2]] = m[1].toLowerCase();
  }
  return out;
}

async function downloadBinary() {
  if (checkPlatformPackage()) {
    return 0;
  }

  const os = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!os || !arch) {
    console.warn(
      `[notion-cli] Unsupported platform: ${process.platform}-${process.arch}`
    );
    console.warn("[notion-cli] You may need to build from source.");
    return 0;
  }

  const version = getVersion();
  const ext = process.platform === "win32" ? ".exe" : "";
  const binaryName = `notion-cli-${os}-${arch}${ext}`;
  const base = `https://github.com/Coastal-Programs/notion-cli/releases/download/v${version}`;
  const checksumURL = `${base}/SHA256SUMS`;
  const binaryURL = `${base}/${binaryName}`;

  const destDir = path.join(
    __dirname,
    "node_modules",
    ".cache",
    "notion-cli",
    "bin"
  );
  const destPath = path.join(
    destDir,
    process.platform === "win32" ? "notion-cli.exe" : "notion-cli"
  );

  try {
    fs.mkdirSync(destDir, { recursive: true });
  } catch {
    // Directory may already exist
  }

  console.log(`[notion-cli] Downloading binary for ${os}/${arch}...`);

  // ---- Step 1: fetch the signed checksum manifest ----
  let checksums;
  try {
    const sums = await fetchText(checksumURL);
    if (!sums.ok) {
      console.warn(
        `[notion-cli] SHA256SUMS not available (HTTP ${sums.status}).`
      );
      console.warn(
        "[notion-cli] Skipping postinstall download. Build from source: make build"
      );
      return 0;
    }
    checksums = parseChecksums(sums.body);
  } catch (err) {
    console.warn(`[notion-cli] Failed to fetch SHA256SUMS: ${err.message}`);
    console.warn("[notion-cli] You can build from source: make build");
    return 0;
  }

  const expected = checksums[binaryName];
  if (!expected) {
    console.warn(
      `[notion-cli] No checksum entry for ${binaryName} in SHA256SUMS.`
    );
    console.warn(
      "[notion-cli] Refusing to install an unverified binary. Build from source: make build"
    );
    return 0;
  }

  // ---- Step 2: stream the binary, hashing as we go, into a temp file ----
  const tmpPath = `${destPath}.partial`;
  // Clean any stale partial from a previous failed run.
  try {
    fs.unlinkSync(tmpPath);
  } catch {
    // not present
  }
  const file = fs.createWriteStream(tmpPath);
  const hash = crypto.createHash("sha256");

  let downloadResult;
  try {
    downloadResult = await fetchURL(binaryURL, (chunk) => {
      hash.update(chunk);
      file.write(chunk);
    });
  } catch (err) {
    try {
      file.destroy();
    } catch {
      // best-effort
    }
    try {
      fs.unlinkSync(tmpPath);
    } catch {
      // best-effort
    }
    console.warn(`[notion-cli] Download failed: ${err.message}`);
    console.warn("[notion-cli] You can build from source: make build");
    return 0;
  }

  // Wait for the file stream to flush before checking the hash.
  await new Promise((resolve, reject) => {
    file.end(() => resolve());
    file.on("error", reject);
  });

  if (!downloadResult.ok) {
    try {
      fs.unlinkSync(tmpPath);
    } catch {
      // best-effort
    }
    console.warn(
      `[notion-cli] Binary not available (HTTP ${downloadResult.status}).`
    );
    console.warn(
      "[notion-cli] You may need to build from source: make build"
    );
    return 0;
  }

  const actual = hash.digest("hex");
  if (actual !== expected) {
    try {
      fs.unlinkSync(tmpPath);
    } catch {
      // best-effort
    }
    // This is an active-attack signal, not a transient network failure.
    // Hard-fail: never persist a binary whose hash does not match.
    console.error(
      `[notion-cli] CHECKSUM MISMATCH for ${binaryName}: expected ${expected}, got ${actual}`
    );
    console.error(
      "[notion-cli] Refusing to install. Report this at https://github.com/Coastal-Programs/notion-cli/issues"
    );
    return 1;
  }

  // ---- Step 3: atomic rename + chmod ----
  try {
    fs.renameSync(tmpPath, destPath);
    if (process.platform !== "win32") {
      fs.chmodSync(destPath, 0o755);
    }
  } catch (err) {
    try {
      fs.unlinkSync(tmpPath);
    } catch {
      // best-effort
    }
    console.warn(`[notion-cli] Failed to finalize binary: ${err.message}`);
    return 0;
  }

  console.log("[notion-cli] Binary installed and checksum-verified.");
  return 0;
}

downloadBinary()
  .then((code) => {
    if (code !== 0) {
      process.exit(code);
    }
  })
  .catch(() => {
    // Unexpected error path: do not break npm install for transient issues.
  });
