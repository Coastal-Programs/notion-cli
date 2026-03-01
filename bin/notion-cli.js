#!/usr/bin/env node

"use strict";

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const PLATFORM_PACKAGES = {
  "darwin-arm64": "@coastal-programs/notion-cli-darwin-arm64",
  "darwin-x64": "@coastal-programs/notion-cli-darwin-x64",
  "linux-x64": "@coastal-programs/notion-cli-linux-x64",
  "linux-arm64": "@coastal-programs/notion-cli-linux-arm64",
  "win32-x64": "@coastal-programs/notion-cli-win32-x64",
};

function getBinaryPath() {
  const platformKey = `${process.platform}-${process.arch}`;
  const pkg = PLATFORM_PACKAGES[platformKey];

  if (!pkg) {
    console.error(
      `Unsupported platform: ${process.platform}-${process.arch}\n` +
        `Supported: ${Object.keys(PLATFORM_PACKAGES).join(", ")}`
    );
    process.exit(1);
  }

  // Try platform-specific optional dependency first
  try {
    const pkgPath = require.resolve(`${pkg}/package.json`);
    const pkgDir = path.dirname(pkgPath);
    const ext = process.platform === "win32" ? ".exe" : "";
    const binPath = path.join(pkgDir, "bin", `notion-cli${ext}`);
    if (fs.existsSync(binPath)) {
      return binPath;
    }
  } catch {
    // Optional dependency not installed, try fallback
  }

  // Fallback: binary downloaded by postinstall
  const cacheBin = path.join(
    __dirname,
    "..",
    "node_modules",
    ".cache",
    "notion-cli",
    "bin",
    process.platform === "win32" ? "notion-cli.exe" : "notion-cli"
  );
  if (fs.existsSync(cacheBin)) {
    return cacheBin;
  }

  // Fallback: global install location
  const homeBin = path.join(
    process.env.HOME || process.env.USERPROFILE || "",
    ".notion-cli",
    "bin",
    process.platform === "win32" ? "notion-cli.exe" : "notion-cli"
  );
  if (fs.existsSync(homeBin)) {
    return homeBin;
  }

  console.error(
    `notion-cli binary not found for ${platformKey}.\n` +
      `Try reinstalling: npm install -g @coastal-programs/notion-cli`
  );
  process.exit(1);
}

const binPath = getBinaryPath();

try {
  const result = execFileSync(binPath, process.argv.slice(2), {
    stdio: "inherit",
    env: process.env,
  });
} catch (err) {
  if (err.status !== null) {
    process.exit(err.status);
  }
  process.exit(1);
}
