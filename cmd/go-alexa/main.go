// Package main runs the go-alexa terminal client.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/portpowered/go-alexa/cmd/go-alexa/internal/cli"
)

const canceledExitCode = 130

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	application := cli.New(os.Stdin, os.Stdout, os.Stderr)

	err := application.Run(ctx, os.Args[1:])
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return canceledExitCode
		}

		_, _ = fmt.Fprintf(os.Stderr, "go-alexa: %s\n", err)

		return 1
	}

	return 0
}
