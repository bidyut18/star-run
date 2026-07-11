package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// PackageManager represents the detected package manager
type PackageManager int

const (
	Npm PackageManager = iota
	Yarn
	Pnpm
	Bun
)

func (pm PackageManager) String() string {
	switch pm {
	case Npm:
		return "npm"
	case Yarn:
		return "yarn"
	case Pnpm:
		return "pnpm"
	case Bun:
		return "bun"
	default:
		return "unknown"
	}
}

func (pm PackageManager) installArgs() []string {
	return []string{"install"}
}

func (pm PackageManager) runArgs(script string) []string {
	if pm == Yarn {
		return []string{script} // Yarn uses `yarn script`
	}
	return []string{"run", script} // Others use `pm run script`
}

// detectPackageManager detects the package manager based on lockfiles
// and the packageManager field in package.json (traversing upward).
// If stopDir is non-empty, traversal stops at that directory (for testing).
func detectPackageManager(dir, stopDir string) (PackageManager, error) {
	for {
		// 1. Check package.json for a "packageManager" hint
		pkgPath := filepath.Join(dir, "package.json")
		if data, err := os.ReadFile(pkgPath); err == nil {
			var pkg struct {
				PackageManager string `json:"packageManager"`
			}
			if err := json.Unmarshal(data, &pkg); err == nil && pkg.PackageManager != "" {
				name := pkg.PackageManager
				if idx := strings.Index(name, "@"); idx != -1 {
					name = name[:idx]
				}
				switch name {
				case "npm":
					return Npm, nil
				case "yarn":
					return Yarn, nil
				case "pnpm":
					return Pnpm, nil
				case "bun":
					return Bun, nil
				}
				// unknown – fall through to lockfiles
			}
		}

		// 2. Check for lockfiles
		if _, err := os.Stat(filepath.Join(dir, "package-lock.json")); err == nil {
			return Npm, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); err == nil {
			return Yarn, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); err == nil {
			return Pnpm, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "bun.lock")); err == nil {
			return Bun, nil
		}
		if _, err := os.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
			return Bun, nil
		}

		// 3. Stop if we've reached the stopDir (if set) and found nothing
		if stopDir != "" && dir == stopDir {
			break
		}

		// 4. Move up one level
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return -1, errors.New("no package manager detected")
}