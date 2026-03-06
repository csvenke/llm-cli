package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"llm/internal/cmd"
	"llm/internal/providers"
)

var version string

func main() {
	var printVersion bool
	flag.BoolVar(&printVersion, "v", false, "print version")
	flag.BoolVar(&printVersion, "version", false, "print version")
	flag.Usage = usage
	flag.Parse()

	if printVersion {
		if version == "" {
			version = "snapshot"
		}
		fmt.Println(version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	// Create a context that can be cancelled via signals (Ctrl+C, SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	provider, err := providers.ResolveByAPIKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = cmd.Run(ctx, provider, os.Stdout, os.Stderr, args)
	if errors.Is(err, cmd.ErrUnknownCommand) {
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	usageTo(os.Stderr)
}

func usageTo(w io.Writer) {
	cmd.UsageTo(w)
	_, _ = fmt.Fprintf(w, "\nOptions:\n")
	oldOutput := flag.CommandLine.Output()
	flag.CommandLine.SetOutput(w)
	defer flag.CommandLine.SetOutput(oldOutput)
	flag.PrintDefaults()
}
