import { fileURLToPath } from "url";
import * as fs from "fs";
import * as path from "path";
import * as crypto from "crypto";
import rootPkg from "../package.json";

const NPM_SCOPE = "@bidyut26";
const VERSION: string = rootPkg.version;

interface Target {
  platform: string;
  arch: string;
  binDir: string;
  bin: string;
}

const TARGETS: Target[] = [
  { platform: "darwin", arch: "x64", binDir: "darwin-x64", bin: "star-run" },
  { platform: "darwin", arch: "arm64", binDir: "darwin-arm64", bin: "star-run" },
  { platform: "linux", arch: "x64", binDir: "linux-x64", bin: "star-run" },
  { platform: "linux", arch: "arm64", binDir: "linux-arm64", bin: "star-run" },
  { platform: "win32", arch: "x64", binDir: "win32-x64", bin: "star-run.exe" },
];

const ROOT_DIR = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const NPM_DIR = path.join(ROOT_DIR, "npm");
const BIN_DIR = path.join(ROOT_DIR, "bin");

function rmrf(dir: string): void {
  if (fs.existsSync(dir)) fs.rmSync(dir, { recursive: true, force: true });
}

function writeJson(file: string, data: unknown): void {
  fs.writeFileSync(file, JSON.stringify(data, null, 2) + "\n");
}

function copyExecutable(src: string, dst: string): void {
  fs.copyFileSync(src, dst);
  if (process.platform !== "win32") {
    fs.chmodSync(dst, 0o755);
  }
}

function sha256(file: string): string {
  return crypto.createHash("sha256").update(fs.readFileSync(file)).digest("hex");
}

rmrf(NPM_DIR);
fs.mkdirSync(NPM_DIR, { recursive: true });

const optionalDeps: Record<string, string> = {};
let missing = 0;

for (const t of TARGETS) {
  const pkgName = `${NPM_SCOPE}/star-run-${t.platform}-${t.arch}`;
  const dirName = `star-run-${t.platform}-${t.arch}`;
  const pkgDir = path.join(NPM_DIR, dirName);
  fs.mkdirSync(pkgDir, { recursive: true });

  const srcBin = path.join(BIN_DIR, t.binDir, t.bin);
  const dstBin = path.join(pkgDir, t.bin);

  if (!fs.existsSync(srcBin)) {
    console.warn(`⚠️  Missing binary: ${t.binDir}/${t.bin}`);
    missing++;
    continue;
  }

  copyExecutable(srcBin, dstBin);
  const checksum = sha256(dstBin);

  // FIX: Add repository.directory for each platform package
  writeJson(path.join(pkgDir, "package.json"), {
    name: pkgName,
    version: VERSION,
    description: `Platform-specific binary for ${rootPkg.name} on ${t.platform} ${t.arch}`,
    author: rootPkg.author,
    license: rootPkg.license,
    os: [t.platform],
    cpu: [t.arch],
    files: [t.bin, "README.md"],
    repository: {
      ...rootPkg.repository,
      directory: `npm/${dirName}`,
    },
    homepage: rootPkg.homepage,
    bugs: rootPkg.bugs,
    publishConfig: { access: "public" },
    starRunChecksum: checksum,
  });

  fs.writeFileSync(
    path.join(pkgDir, "README.md"),
    `# ${pkgName}\n\nPlatform-specific binary for [${rootPkg.name}](${rootPkg.homepage}) on ${t.platform} ${t.arch}.\n`,
  );

  optionalDeps[pkgName] = VERSION;
  console.log(`✅  ${pkgName}`);
}

if (missing > 0) {
  console.error(`\n❌ ${missing} binary(s) missing. Run: task build-all`);
  process.exit(1);
}

const mainDir = path.join(NPM_DIR, "star-run");
fs.mkdirSync(mainDir, { recursive: true });

const srcIndex = path.join(ROOT_DIR, "dist", "index.mjs");
if (!fs.existsSync(srcIndex)) {
  console.error("❌ dist/index.mjs not found. Run: npm run build:ts");
  process.exit(1);
}
fs.copyFileSync(srcIndex, path.join(mainDir, "index.mjs"));
if (process.platform !== "win32") {
  fs.chmodSync(path.join(mainDir, "index.mjs"), 0o755);
}

for (const file of ["README.md", "LICENSE"]) {
  const src = path.join(ROOT_DIR, file);
  if (fs.existsSync(src)) fs.copyFileSync(src, path.join(mainDir, file));
}


writeJson(path.join(mainDir, "package.json"), {
  name: rootPkg.name,
  version: rootPkg.version,
  description: rootPkg.description,
  main: "index.mjs",
  bin: { "star-run": "index.mjs" },
  files: ["index.mjs", "README.md", "LICENSE"],
  optionalDependencies: optionalDeps,
  keywords: rootPkg.keywords,
  author: rootPkg.author,
  license: rootPkg.license,
  repository: {
    ...rootPkg.repository,
    directory: "npm/star-run",
  },
  bugs: rootPkg.bugs,
  homepage: rootPkg.homepage,
  engines: rootPkg.engines,
  publishConfig: { access: "public" },
});

console.log(`✅  ${rootPkg.name} wrapper (v${VERSION})`);
console.log("\n📦 Publish order: platform packages → star-run wrapper");