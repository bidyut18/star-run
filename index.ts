#!/usr/bin/env node
import { spawn } from "child_process";
import * as path from "path";
import * as fs from "fs";
import * as os from "os";
import { fileURLToPath } from "url";
import { createRequire } from "module";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);

const PLATFORM = os.platform();
const ARCH = os.arch();
const BINARY_NAME = PLATFORM === "win32" ? "star-run.exe" : "star-run";
const NPM_SCOPE = "@bidyut26";

function getBinaryPath(): string | null {
  const platformPkg = `${NPM_SCOPE}/star-run-${PLATFORM}-${ARCH}`;
  try {
    const pkgPath = require.resolve(`${platformPkg}/package.json`);
    const binaryPath = path.join(path.dirname(pkgPath), BINARY_NAME);
    if (fs.existsSync(binaryPath)) return binaryPath;
  } catch {
    
  }

  const candidates = [
    path.join(__dirname, "bin", BINARY_NAME),
    path.join(__dirname, "..", "bin", BINARY_NAME),
    path.join(process.cwd(), "bin", BINARY_NAME),
  ];

  for (const p of candidates) {
    if (fs.existsSync(p)) return p;
  }

  return null;
}

const binaryPath = getBinaryPath();

if (!binaryPath) {
  console.error("");
  console.error(`❌  star-run: No binary found for ${PLATFORM}-${ARCH}.`);
  console.error("");
  console.error(`   Supported platforms:`);
  console.error(`   • macOS:   darwin-x64, darwin-arm64`);
  console.error(`   • Linux:   linux-x64, linux-arm64`);
  console.error(`   • Windows: win32-x64`);
  console.error("");
  console.error(`   Install from source:`);
  console.error(`   git clone https://github.com/bidyut18/star-run.git`);
  console.error(`   cd star-run && go build -o bin/star-run ./src`);
  console.error("");
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: true,
});

child.on("exit", (code: number | null) => process.exit(code ?? 0));
child.on("error", (err: Error) => {
  console.error(`❌  Failed to spawn star-run: ${err.message}`);
  process.exit(1);
});