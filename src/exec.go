package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func runCommand(ctx context.Context, name string, args []string) int {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.Canceled {
			return 130
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}

		fatalf("Failed to execute '%s': %v\n(Is it installed and in your PATH?)\n", name, err)
	}
	return 0
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}