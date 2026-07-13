package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	installFlag := flag.Bool("install", false, "Install dependencies using the detected package manager")
	listFlag := flag.Bool("list", false, "List available scripts from package.json")
	detectFlag := flag.Bool("detect", false, "Show the detected package manager")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "star-run: Universal package manager script runner\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  star-run <script> [args...]\n")
		fmt.Fprintf(os.Stderr, "  star-run --install\n")
		fmt.Fprintf(os.Stderr, "  star-run --list\n")
		fmt.Fprintf(os.Stderr, "  star-run --detect\n")
	}
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fatalf("Failed to get working directory: %v\n", err)
	}

	pm, err := detectPackageManager(cwd, "")
	if err != nil {
		fatalf("Detection Error: No lockfile or packageManager field found in tree.\n")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch {
	case *installFlag:
		fmt.Printf("📦 Installing dependencies via %s...\n", pm)
		runCommand(ctx, pm.String(), pm.installArgs())
	case *listFlag:
		listScripts(cwd, pm, "")
	case *detectFlag:
		fmt.Println(pm)
	default:
		args := flag.Args()
		if len(args) == 0 {
			flag.Usage()
			os.Exit(1)
		}

		cmdArgs := append(pm.runArgs(args[0]), args[1:]...)
		fmt.Printf("🚀 Executing: %s %s\n", pm, strings.Join(cmdArgs, " "))
		runCommand(ctx, pm.String(), cmdArgs)
	}
}
