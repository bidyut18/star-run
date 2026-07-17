package main

import "testing"

func TestPackageManagerRunArgs(t *testing.T) {
	tests := []struct {
		pm       PackageManager
		script   string
		expected []string
	}{
		{Npm, "test", []string{"run", "test"}},
		{Yarn, "test", []string{"test"}},
		{Pnpm, "test", []string{"run", "test"}},
		{Bun, "test", []string{"run", "test"}},
		{Deno, "test", []string{"task", "test"}},
	}
	for _, tt := range tests {
		t.Run(string(tt.pm), func(t *testing.T) {
			args := tt.pm.runArgs(tt.script)
			if len(args) != len(tt.expected) {
				t.Fatalf("%v: expected %v, got %v", tt.pm, tt.expected, args)
			}
			for i := range args {
				if args[i] != tt.expected[i] {
					t.Errorf("%v: expected %v, got %v", tt.pm, tt.expected, args)
				}
			}
		})
	}
}

func TestPackageManagerInstallArgs(t *testing.T) {
	for _, pm := range []PackageManager{Npm, Yarn, Pnpm, Bun, Deno} { 
		t.Run(string(pm), func(t *testing.T) {
			args := pm.installArgs()
			if len(args) != 1 || args[0] != "install" {
				t.Errorf("%v: expected [install], got %v", pm, args)
			}
		})
	}
}