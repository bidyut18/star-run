package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func runCommand(ctx context.Context, name string, args []string) {
	// Bind command execution to the OS signal context
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Differentiate between OS interrupt termination and real execution errors
		if ctx.Err() == context.Canceled {
			os.Exit(130) // Standard exit code for SIGINT
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		fatalf("Failed to execute '%s': %v\n(Is it installed and in your PATH?)\n", name, err)
	}
	os.Exit(0)
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}