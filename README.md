# 🚀 cat-run

> **Universal package manager script runner** – fast, zero‑config, and distributed globally via npm.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![npm version](https://img.shields.io/npm/v/cat-run)](https://www.npmjs.com/package/cat-run)

Stop trying to remember whether the current repository uses `npm`, `yarn`, `pnpm`, or `bun`. Just type `cat-run dev` and let the runner instantly figure it out for you.

---

## ✨ Features

- ⚡️ **Near‑instant startup** – written in Go, boots in ~3–5 ms with zero runtime overhead.
- 🧠 **Smart detection** – automatically traverses parent directories to identify the correct environment.
- 📡 **Signal forwarding** – gracefully handles `SIGINT`/`SIGTERM` to safely shut down running scripts.
- 📦 **Trivial distribution** – installs as a standalone static binary natively via npm.

---

## 📦 Installation

```bash
npm install -g cat-run
```

## Usage

### Run any script defined in package.json

cat-run dev
cat-run build --watch

### Install dependencies using the detected package manager

cat-run --install

### List available scripts beautifully formatted

cat-run --list

### Show the detected package manager without executing anything

cat-run --detect
