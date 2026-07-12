#!/usr/bin/env node
/**
 * cat-run launcher
 * Resolves the platform-specific binary from optionalDependencies,
 * falls back to local bin/ for development.
 */

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const os = require("os");

const PLATFORM = os.platform();
const ARCH = os.arch();
const BINARY_NAME = PLATFORM === "win32" ? "cat-run.exe" : "cat-run";
const NPM_SCOPE = "@bidyut26";

function getBinaryPath() {
  // 1. Try platform-specific optional dependency
  const platformPkg = `${NPM_SCOPE}/cat-run-${PLATFORM}-${ARCH}`;
  try {
    const pkgPath = require.resolve(`${platformPkg}/package.json`);
    const binaryPath = path.join(path.dirname(pkgPath), BINARY_NAME);
    if (fs.existsSync(binaryPath)) return binaryPath;
  } catch {
    // Not installed for this platform — that's fine, it's optional
  }

  // 2. Fallback: local bin/ (development / manual install)
  const localBin = path.join(__dirname, "bin", BINARY_NAME);
  if (fs.existsSync(localBin)) return localBin;

  // 3. Fallback: same directory as this script (development)
  const devBin = path.join(__dirname, BINARY_NAME);
  if (fs.existsSync(devBin)) return devBin;

  return null;
}

const binaryPath = getBinaryPath();

if (!binaryPath) {
  console.error(``);
  console.error(`❌  cat-run: No binary found for ${PLATFORM}-${ARCH}.`);
  console.error(``);
  console.error(`   Supported platforms:`);
  console.error(`   • macOS:   darwin-x64, darwin-arm64`);
  console.error(`   • Linux:   linux-x64, linux-arm64`);
  console.error(`   • Windows: win32-x64`);
  console.error(``);
  console.error(`   Install from source:`);
  console.error(`   git clone https://github.com/bidyut18/cat-run.git`);
  console.error(`   cd cat-run && go build -o bin/cat-run ./src`);
  console.error(``);
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: true,
});

child.on("exit", (code) => process.exit(code ?? 0));
child.on("error", (err) => {
  console.error(`❌  Failed to spawn cat-run: ${err.message}`);
  process.exit(1);
});