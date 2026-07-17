package main

import (
	"bytes"
	"encoding/json"
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

func TestDenoConfigReader(t *testing.T) {
	tmpDir := t.TempDir()
	content := `{
  "tasks": {
    "dev": "deno run --watch main.ts",
    "build": "deno compile main.ts"
  }
}`
	cfgPath := filepath.Join(tmpDir, "deno.json")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	reader := DenoConfigReader{}
	scripts, err := reader.ReadScripts(cfgPath)
	if err != nil {
		t.Fatalf("ReadScripts failed: %v", err)
	}
	if len(scripts) != 2 {
		t.Errorf("expected 2 scripts, got %d", len(scripts))
	}
	// Check that "dev" exists
	found := false
	for _, s := range scripts {
		if s.Name == "dev" && s.Command == "deno run --watch main.ts" {
			found = true
		}
	}
	if !found {
		t.Error("script 'dev' not found")
	}

	has, err := reader.HasScript(cfgPath, "build")
	if err != nil || !has {
		t.Errorf("HasScript build: expected true, got %v (err: %v)", has, err)
	}
	has, err = reader.HasScript(cfgPath, "missing")
	if err != nil || has {
		t.Errorf("HasScript missing: expected false, got %v", has)
	}
}

func TestLocateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"scripts":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	// Create a deno.json in a subdirectory
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	denoPath := filepath.Join(subDir, "deno.json")
	if err := os.WriteFile(denoPath, []byte(`{"tasks":{}}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Start in subDir, with Deno package manager – should find deno.json
	cfg, err := LocateConfig(subDir, tmpDir, Deno)
	if err != nil {
		t.Fatalf("LocateConfig failed: %v", err)
	}
	if cfg.Path != denoPath || cfg.Type != ConfigDenoJSON {
		t.Errorf("expected deno.json, got %s (type %v)", cfg.Path, cfg.Type)
	}

	// Start in subDir with npm – should find package.json in parent
	cfg, err = LocateConfig(subDir, tmpDir, Npm)
	if err != nil {
		t.Fatalf("LocateConfig failed: %v", err)
	}
	if cfg.Path != pkgPath || cfg.Type != ConfigPackageJSON {
		t.Errorf("expected package.json, got %s (type %v)", cfg.Path, cfg.Type)
	}
}

func TestListScriptsDeno(t *testing.T) {
	tmpDir := t.TempDir()
	content := `{
  "tasks": {
    "start": "deno run index.ts"
  }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "deno.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	svc := ScriptService{
		Renderer: ScriptRenderer{Writer: &buf},
	}
	if err := svc.ListScripts(tmpDir, Deno, tmpDir); err != nil {
		t.Fatalf("ListScripts failed: %v", err)
	}
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("start")) {
		t.Errorf("output missing script 'start': %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("deno run index.ts")) {
		t.Errorf("output missing command: %s", output)
	}
}

func TestStripJSONC(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"url in string", `{"url":"http://a.com"}`, `{"url":"http://a.com"}`, false},
		{"comment in string", `{"a":"//b"}`, `{"a":"//b"}`, false},
		{"line comment", `{"a":1}`, `{"a":1}`, false},
		{"block comment inline", `{"a":/*c*/1}`, `{"a":1}`, false},
		{"block comment multiline", `{
/* comment */
"a":1
}`, `{

"a":1
}`, false},
		{"trailing comma", `{"a":1,}`, `{"a":1,}`, false},
		{"unterminated string", `{"a":"`, ``, true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stripJSONC([]byte(tt.in))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("stripJSONC(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDetectFreshProject(t *testing.T) {
	t.Run("only package.json no lockfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test"}`), 0644); err != nil {
			t.Fatal(err)
		}
		pm, warning, err := detectPackageManager(tmpDir, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pm != Npm {
			t.Errorf("expected npm, got %v", pm)
		}
		if warning != "" {
			t.Errorf("expected no warning, got %q", warning)
		}
	})

	t.Run("only deno.json no deno.lock", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmpDir, "deno.json"), []byte(`{"tasks":{}}`), 0644); err != nil {
			t.Fatal(err)
		}
		pm, warning, err := detectPackageManager(tmpDir, tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pm != Deno {
			t.Errorf("expected deno, got %v", pm)
		}
		if warning != "" {
			t.Errorf("expected no warning, got %q", warning)
		}
	})
}

func TestReadConfigFileTrailingComma(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "deno.jsonc")
	content := `{"tasks": {"start": "deno run index.ts",}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := readConfigFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var cfg struct {
		Tasks map[string]string `json:"tasks"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse cleaned JSON: %v", err)
	}
	if _, ok := cfg.Tasks["start"]; !ok {
		t.Error("expected task 'start' to exist")
	}
}
