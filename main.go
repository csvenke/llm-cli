package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"llm/internal/ask"
	"llm/internal/commit"
	"llm/internal/git"
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

	switch args[0] {
	case "ask":
		err = ask.Run(ctx, provider, os.Stdout, args[1:])
	case "commit":
		err = commit.Run(ctx, provider, &git.RealClient{}, os.Stderr, args[1:])
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: llm <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  ask <question>      Ask a question\n")
	fmt.Fprintf(os.Stderr, "  commit [-a|--amend] Draft a commit message\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}
