package main

import (
	"context"
	"fmt"
	"os"

	"code_agent/internal/app"
)

func main() {
	ctx := context.Background()

	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init app: %v\n", err)
		os.Exit(1)
	}

	if err := application.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "code-agent: %v\n", err)
		os.Exit(1)
	}
}
