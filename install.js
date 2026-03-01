#!/usr/bin/env node

"use strict";

const https = require("https");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  arm64: "arm64",
  x64: "amd64",
};

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

async function downloadBinary() {
  if (checkPlatformPackage()) {
    return;
  }

  const os = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!os || !arch) {
    console.warn(
      `[notion-cli] Unsupported platform: ${process.platform}-${process.arch}`
    );
    console.warn("[notion-cli] You may need to build from source.");
    return;
  }

  const version = getVersion();
  const ext = process.platform === "win32" ? ".exe" : "";
  const binaryName = `notion-cli-${os}-${arch}${ext}`;
  const url = `https://github.com/Coastal-Programs/notion-cli/releases/download/v${version}/${binaryName}`;

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

  return new Promise((resolve) => {
    const download = (downloadUrl, redirects = 0) => {
      if (redirects > 5) {
        console.warn("[notion-cli] Too many redirects. Skipping download.");
        resolve();
        return;
      }

      https
        .get(downloadUrl, { headers: { "User-Agent": "notion-cli-npm" } }, (res) => {
          if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
            download(res.headers.location, redirects + 1);
            return;
          }

          if (res.statusCode !== 200) {
            console.warn(
              `[notion-cli] Binary not available (HTTP ${res.statusCode}).`
            );
            console.warn(
              "[notion-cli] You may need to build from source: make build"
            );
            res.resume();
            resolve();
            return;
          }

          const file = fs.createWriteStream(destPath);
          res.pipe(file);
          file.on("finish", () => {
            file.close();
            if (process.platform !== "win32") {
              fs.chmodSync(destPath, 0o755);
            }
            console.log("[notion-cli] Binary installed successfully.");
            resolve();
          });
        })
        .on("error", (err) => {
          console.warn(`[notion-cli] Download failed: ${err.message}`);
          console.warn("[notion-cli] You can build from source: make build");
          resolve();
        });
    };

    download(url);
  });
}

downloadBinary().catch(() => {
  // Silently fail - postinstall should never break npm install
});
