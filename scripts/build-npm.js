#!/usr/bin/env node
/**
 * Package Go binaries into npm platform-specific tarballs.
 *
 * Prerequisites:
 *   make build-all
 *
 * Usage:
 *   node scripts/build-npm.js
 *
 * Publish order:
 *   1. Platform packages first
 *   2. Main package last
 */

const fs = require("fs");
const path = require("path");

// ----- version from root -----
const rootPkg = require("../package.json");
const version = rootPkg.version;

const author = "Bidyut Mahanta <bidyutmahanta7768@outlook.com>";
const repoUrl = "https://github.com/yourname/uni-run";

const targets = [
  { platform: "darwin", arch: "x64", binDir: "darwin-amd64" },
  { platform: "darwin", arch: "arm64", binDir: "darwin-arm64" },
  { platform: "linux", arch: "x64", binDir: "linux-amd64" },
  { platform: "linux", arch: "arm64", binDir: "linux-arm64" },
  { platform: "win32", arch: "x64", binDir: "windows-amd64" },
];

const npmDir = path.join(__dirname, "..", "npm");
const binDir = path.join(__dirname, "..", "bin");

let missingCount = 0;

// ---- 1. Create platform packages ----
for (const target of targets) {
  const pkgName = `uni-run-${target.platform}-${target.arch}`;
  const pkgDir = path.join(npmDir, pkgName);
  fs.mkdirSync(pkgDir, { recursive: true });

  const binaryName = target.platform === "win32" ? "uni-run.exe" : "uni-run";
  const srcBinary = path.join(binDir, target.binDir, binaryName);
  const destBinary = path.join(pkgDir, binaryName);

  if (!fs.existsSync(srcBinary)) {
    console.warn(
      `⚠️  Missing: ${target.binDir}/${binaryName} (run: make build-all)`,
    );
    missingCount++;
    continue;
  }

  fs.copyFileSync(srcBinary, destBinary);
  fs.chmodSync(destBinary, 0o755);

  const pkgJson = {
    name: pkgName,
    version: version, // use root version
    description: `uni-run binary for ${target.platform}-${target.arch}`,
    author,
    os: [target.platform],
    cpu: [target.arch],
    files: [binaryName],
    license: "MIT",
    repository: {
      type: "git",
      url: `${repoUrl}.git`,
    },
    homepage: repoUrl,
  };

  fs.writeFileSync(
    path.join(pkgDir, "package.json"),
    JSON.stringify(pkgJson, null, 2) + "\n",
  );

  console.log(`✅  Created ${pkgName} (v${version})`);
}

if (missingCount > 0) {
  console.warn(`\n⚠️  ${missingCount} binary(s) missing. Run: make build-all`);
  process.exit(1);
}

// ---- 2. Create main wrapper package ----
function createMainWrapper() {
  const mainDir = path.join(npmDir, "uni-run");
  fs.mkdirSync(mainDir, { recursive: true });

  // (Optional) copy index.js and bin folder if not present
  // but we assume they exist in the repo already.

  const pkgJson = {
    name: "uni-run",
    version: version,
    description:
      "Universal package manager script runner — fast Go binary distributed via npm",
    main: "index.js",
    bin: { "uni-run": "index.js" },
    files: ["index.js", "bin", "README.md", "LICENSE"],
    optionalDependencies: {
      "uni-run-darwin-x64": version,
      "uni-run-darwin-arm64": version,
      "uni-run-linux-x64": version,
      "uni-run-linux-arm64": version,
      "uni-run-win32-x64": version,
    },
    keywords: [
      "cli",
      "package-manager",
      "npm",
      "yarn",
      "pnpm",
      "bun",
      "runner",
      "go",
    ],
    author: author,
    license: "MIT",
    repository: {
      type: "git",
      url: `${repoUrl}.git`,
    },
    bugs: { url: `${repoUrl}/issues` },
    homepage: `${repoUrl}#readme`,
    engines: { node: ">=16" },
  };

  fs.writeFileSync(
    path.join(mainDir, "package.json"),
    JSON.stringify(pkgJson, null, 2) + "\n",
  );

  console.log(`✅  Created main wrapper (v${version})`);
}

createMainWrapper();

console.log("\n✅  Done! Publish order:");
console.log("    1. npm/uni-run-* (platform packages)");
console.log("    2. npm/uni-run (main wrapper)");
