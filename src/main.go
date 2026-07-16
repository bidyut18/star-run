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


var version = "dev"

func main() {
	installFlag := flag.Bool("install", false, "Install dependencies using the detected package manager")
	listFlag := flag.Bool("list", false, "List available scripts from package.json")
	detectFlag := flag.Bool("detect", false, "Show the detected package manager")
	versionFlag := flag.Bool("version", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "star-run %s — Universal package manager script runner\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  star-run <script> [args...]    Run a package.json script\n")
		fmt.Fprintf(os.Stderr, "  star-run --install             Install dependencies\n")
		fmt.Fprintf(os.Stderr, "  star-run --list                List available scripts\n")
		fmt.Fprintf(os.Stderr, "  star-run --detect              Show detected package manager\n")
		fmt.Fprintf(os.Stderr, "  star-run --version             Show version\n")
		fmt.Fprintf(os.Stderr, "  star-run --help                Show this help message\n")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  star-run dev\n")
		fmt.Fprintf(os.Stderr, "  star-run build --watch\n")
		fmt.Fprintf(os.Stderr, "  star-run test --coverage\n")
		
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("star-run %s\n", version)
		os.Exit(0)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fatalf("Failed to get working directory: %v\n", err)
	}

	pm, warning, err := detectPackageManager(cwd, "")
	if err != nil {
		fatalf("Detection Error: No lockfile or packageManager field found in tree.\n")
	}
	if warning != "" {
		fmt.Fprintln(os.Stderr, warning)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch {
	case *installFlag:
		fmt.Printf("📦 Installing dependencies via %s...\n", pm)
		code := runCommand(ctx, pm.String(), pm.installArgs())
		cancel()
		os.Exit(code)

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

		scriptName := args[0]

		svc := ScriptService{Locator: PackageLocator{}, Reader: PackageReader{}}
		if err := svc.ValidateScript(cwd, "", scriptName); err != nil {
			fatalf("Error: %v\n", err)
		}

		cmdArgs := append(pm.runArgs(scriptName), args[1:]...)
		fmt.Printf("🚀 Executing: %s %s\n", pm, strings.Join(cmdArgs, " "))
		code := runCommand(ctx, pm.String(), cmdArgs)
		cancel()
		os.Exit(code)
	}
}
