package main

import (
	"fmt"
	"log/slog"
	"os"
)

func SetupLogger() {
	_, debug := os.LookupEnv("DEBUG")

	style, ok := os.LookupEnv("LOG_STYLE")
	if !ok {
		style = "text"
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		AddSource: debug,
	}

	switch style {
	case "text":
		handler = slog.NewTextHandler(os.Stderr, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		fmt.Fprintf(os.Stderr, "Unknown log style \"%s\", defaulting to text\n", style)
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))
}
