package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)


type PackageJSON struct {
	PackageManager string            `json:"packageManager"`
	Scripts        map[string]string `json:"scripts"`
}

func listScripts(startDir string, pm PackageManager, stopDir string) {
	dir := startDir
	var pkgPath string

	for {
		p := filepath.Join(dir, "package.json")
		if _, err := os.Stat(p); err == nil {
			pkgPath = p
			break
		}
		if stopDir != "" && dir == stopDir {
			fatalf("Error: No package.json found in tree.\n")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fatalf("Error: No package.json found in tree.\n")
		}
		dir = parent
	}

	f, err := os.Open(pkgPath)
	if err != nil {
		fatalf("Error opening package.json: %v\n", err)
	}
	defer f.Close()

	var pkg PackageJSON
	if err := json.NewDecoder(f).Decode(&pkg); err != nil {
		fatalf("Error parsing package.json: %v\n", err)
	}

	if len(pkg.Scripts) == 0 {
		fmt.Println("No scripts found in package.json.")
		return
	}

	fmt.Printf("\n📦 Detected: %s\n", pm)
	fmt.Println(strings.Repeat("─", 40))

	names := make([]string, 0, len(pkg.Scripts))
	maxLen := 0
	for name := range pkg.Scripts {
		names = append(names, name)
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("  %-*s  %s\n", maxLen, name, pkg.Scripts[name])
	}
	fmt.Println()
}
