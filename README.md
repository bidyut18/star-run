# 🚀 star-run

> **Universal package manager script runner** – fast, zero‑config, and distributed globally via npm.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![npm version](https://img.shields.io/npm/v/star-run)](https://www.npmjs.com/package/star-run)
![status](https://img.shields.io/badge/status-alpha-orange)

Stop trying to remember whether the current repository uses `npm`, `yarn`, `pnpm`, or `bun`. Just type `star-run dev` and let the runner instantly figure it out for you.

---

## ✨ Features

- ⚡️ **Near‑instant startup** – written in Go, boots in ~3–5 ms with zero runtime overhead.
- 🧠 **Smart detection** – automatically traverses parent directories to identify the correct environment.
- 📡 **Signal forwarding** – gracefully handles `SIGINT`/`SIGTERM` to safely shut down running scripts.
- 📦 **Trivial distribution** – installs as a standalone static binary natively via npm.
- 🚫 **Zero config** – Works out of the box. No configuration files needed.

---

## 📦 Installation

```bash
npm install -g star-run
```
# Note 

> This package is in 🚀 **alpha**. The API can and will change without warning.


## Usage

### Run any script defined in package.json

star-run dev
star-run build --watch

### Install dependencies using the detected package manager

star-run --install

### List available scripts beautifully formatted

star-run --list

### Show the detected package manager without executing anything

star-run --detect


## 📄 License
This project is licensed under the MIT License – see the LICENSE file for details.