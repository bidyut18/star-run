package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name          string
		files         map[string]string
		expected      PackageManager
		expectWarning string
		expectError   bool
	}{
		{
			name: "packageManager field npm",
			files: map[string]string{
				"package.json": `{"packageManager": "npm@10.0.0"}`,
			},
			expected: Npm,
		},
		{
			name: "packageManager field yarn",
			files: map[string]string{
				"package.json": `{"packageManager": "yarn@1.22.19"}`,
			},
			expected: Yarn,
		},
		{
			name: "packageManager field pnpm",
			files: map[string]string{
				"package.json": `{"packageManager": "pnpm@8.0.0"}`,
			},
			expected: Pnpm,
		},
		{
			name: "packageManager field bun",
			files: map[string]string{
				"package.json": `{"packageManager": "bun@1.0.0"}`,
			},
			expected: Bun,
		},
		{
			name: "lockfile npm",
			files: map[string]string{
				"package-lock.json": `{}`,
			},
			expected: Npm,
		},
		{
			name: "lockfile yarn",
			files: map[string]string{
				"yarn.lock": ``,
			},
			expected: Yarn,
		},
		{
			name: "lockfile pnpm",
			files: map[string]string{
				"pnpm-lock.yaml": ``,
			},
			expected: Pnpm,
		},
		{
			name: "lockfile bun (bun.lock)",
			files: map[string]string{
				"bun.lock": ``,
			},
			expected: Bun,
		},
		{
			name: "lockfile bun (bun.lockb)",
			files: map[string]string{
				"bun.lockb": ``,
			},
			expected: Bun,
		},
		{
			name: "packageManager unknown, fallback to lockfile",
			files: map[string]string{
				"package.json":      `{"packageManager": "unknown"}`,
				"package-lock.json": `{}`,
			},
			expected: Npm,
		},
		{
			name:        "no package manager",
			files:       map[string]string{},
			expected:    "",
			expectError: true,
		},
		{
			name: "traverse up to find lockfile",
			files: map[string]string{
				"subdir/package.json": `{}`,
				"package-lock.json":   `{}`,
			},
			expected: Npm,
		},
		// ─── NEW: warning cases ───
		{
			name: "packageManager npm with conflicting yarn.lock",
			files: map[string]string{
				"package.json": `{"packageManager": "npm@10.0.0"}`,
				"yarn.lock":    ``,
			},
			expected:      Npm,
			expectWarning: "⚠️  packageManager says npm, but found yarn.lock",
		},
		{
			name: "multiple lockfiles prefers npm",
			files: map[string]string{
				"package-lock.json": `{}`,
				"yarn.lock":         ``,
			},
			expected:      Npm,
			expectWarning: "⚠️  Multiple lockfiles found (package-lock.json, yarn.lock); using npm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for relPath, content := range tt.files {
				fullPath := filepath.Join(tmpDir, relPath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			startDir := tmpDir
			if _, ok := tt.files["subdir/package.json"]; ok {
				startDir = filepath.Join(tmpDir, "subdir")
			}

			pm, warning, err := detectPackageManager(startDir, tmpDir)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pm != tt.expected {
				t.Errorf("expected PM %v, got %v", tt.expected, pm)
			}
			if warning != tt.expectWarning {
				t.Errorf("expected warning %q, got %q", tt.expectWarning, warning)
			}
		})
	}
}