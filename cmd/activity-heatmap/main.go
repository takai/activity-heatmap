// Command activity-heatmap generates a weekday/hour heatmap from a
// public YouTube channel's live stream archives.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/takai/activity-heatmap/internal/app"
	"github.com/takai/activity-heatmap/internal/channelurl"
)

const (
	exitOK     = 0
	exitError  = 1
	exitUsage  = 2
	binaryName = "activity-heatmap"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet(binaryName, flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		refresh     = fs.Bool("refresh", false, "Rebuild raw data from the YouTube Data API")
		output      = fs.String("output", "outputs", "Base output directory")
		showVersion = fs.Bool("version", false, "Print version information and exit")
	)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: %s [options] <youtube-channel-url>\n\nOptions:\n", binaryName)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return exitOK
		}
		return exitUsage
	}
	if *showVersion {
		fmt.Fprintln(stdout, version)
		return exitOK
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(stderr, "error: missing channel URL")
		fs.Usage()
		return exitUsage
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "error: too many arguments")
		fs.Usage()
		return exitUsage
	}

	if _, err := channelurl.Parse(fs.Arg(0)); err != nil {
		fmt.Fprintf(stderr, "error: unsupported YouTube channel URL: %s\n", fs.Arg(0))
		return exitUsage
	}

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(stderr, "error: YOUTUBE_API_KEY is required")
		return exitError
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := app.Config{
		ChannelURL: fs.Arg(0),
		OutputDir:  *output,
		Refresh:    *refresh,
		APIKey:     apiKey,
		Stdout:     stdout,
	}
	if err := app.Run(ctx, cfg); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitError
	}
	return exitOK
}
