package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	errNoPackageManager = errors.New("no package manager detected")

	// Ordered slice preserves deterministic priority.
	lockFiles = []struct {
		name string
		pm   PackageManager
	}{
		{"package-lock.json", Npm},
		{"yarn.lock", Yarn},
		{"pnpm-lock.yaml", Pnpm},
		{"bun.lock", Bun},
		{"bun.lockb", Bun},
	}
)

// detectPackageManager walks up the directory tree looking for a package manager.
// It returns the detected PM, an optional warning string, and an error.
func detectPackageManager(dir, stopDir string) (PackageManager, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", "", fmt.Errorf("resolving start directory: %w", err)
	}

	var absStop string
	if stopDir != "" {
		absStop, err = filepath.Abs(stopDir)
		if err != nil {
			return "", "", fmt.Errorf("resolving stop directory: %w", err)
		}
	}

	for {
		// 1. Authoritative packageManager field
		if pm, ok := detectFromPackageJSON(absDir); ok {
			if warning := mismatchWarning(absDir, pm); warning != "" {
				return pm, warning, nil
			}
			return pm, "", nil
		}

		// 2. Lockfile fallback
		if pm, lockfiles, ok := detectFromLockFiles(absDir); ok {
			if len(lockfiles) > 1 {
				warning := fmt.Sprintf("⚠️  Multiple lockfiles found (%s); using %s", strings.Join(lockfiles, ", "), pm)
				return pm, warning, nil
			}
			return pm, "", nil
		}

		if absStop != "" && absDir == absStop {
			break
		}

		parent := filepath.Dir(absDir)
		if parent == absDir {
			break
		}
		absDir = parent
	}

	return "", "", errNoPackageManager
}

func detectFromPackageJSON(dir string) (PackageManager, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return "", false
	}

	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil || pkg.PackageManager == "" {
		return "", false
	}

	name, _, _ := strings.Cut(pkg.PackageManager, "@")
	pm := PackageManager(name)
	switch pm {
	case Npm, Yarn, Pnpm, Bun:
		return pm, true
	}
	return "", false
}

func detectFromLockFiles(dir string) (PackageManager, []string, bool) {
	var found []string
	var pm PackageManager
	for _, lf := range lockFiles {
		if _, err := os.Stat(filepath.Join(dir, lf.name)); err == nil {
			if pm == "" {
				pm = lf.pm // keep first match as the choice
			}
			found = append(found, lf.name)
		}
	}
	return pm, found, len(found) > 0
}

// mismatchWarning returns a message if a lockfile exists for a DIFFERENT PM.
func mismatchWarning(dir string, fieldPM PackageManager) string {
	var found []string
	for _, lf := range lockFiles {
		if lf.pm == fieldPM {
			continue // matching lockfile is not a mismatch
		}
		if _, err := os.Stat(filepath.Join(dir, lf.name)); err == nil {
			found = append(found, lf.name)
		}
	}
	if len(found) > 0 {
		return fmt.Sprintf("⚠️  packageManager says %s, but found %s", fieldPM, strings.Join(found, ", "))
	}
	return ""
}