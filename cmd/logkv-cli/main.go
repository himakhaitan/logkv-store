package main

import (
	"fmt"
	"os"

	"github.com/himakhaitan/logkv-store/cli"
	"github.com/himakhaitan/logkv-store/pkg/config"
	"go.uber.org/fx"
)

func main() {
	// Create a new CLI instance
	var cliInstance *cli.CLI

	app := fx.New(
		fx.NopLogger, // Disable fx logs
		config.Module(),
		cli.Module,
		fx.Populate(&cliInstance),
	)

	// Initialize the application
	if err := app.Err(); err != nil {
		panic(err)
	}

	// Run the CLI with command line arguments
	if err := cliInstance.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
