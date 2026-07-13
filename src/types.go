package main

import (
	"fmt"
)

// PackageManager identifies the package manager.
type PackageManager string

const (
	Npm  PackageManager = "npm"
	Yarn PackageManager = "yarn"
	Pnpm PackageManager = "pnpm"
	Bun  PackageManager = "bun"
)

func (pm PackageManager) String() string {
	switch pm {
	case Npm, Yarn, Pnpm, Bun:
		return string(pm)
	default:
		return "unknown"
	}
}

func (pm PackageManager) installArgs() []string {
	return []string{"install"}
}

func (pm PackageManager) runArgs(script string) []string {
	if pm == Yarn {
		return []string{script}
	}
	return []string{"run", script}
}

// PackageJSON represents the relevant fields of a package.json file.
type PackageJSON struct {
	PackageManager string            `json:"packageManager"`
	Scripts        map[string]string `json:"scripts"`
}

// Script represents a single entry from the "scripts" object.
type Script struct {
	Name    string
	Command string
}

// ErrPackageNotFound is returned when the directory walk ends without finding package.json.
var ErrPackageNotFound = fmt.Errorf("no package.json found in tree")
