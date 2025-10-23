package main

import (
	"log"

	"metrics-scrapper/cmd/internal/cli"
)

func main() {
	rootCmd, err := cli.NewRootCmd()
	if err != nil {
		log.Fatalf("failed to create root cmd: %s", err.Error())
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to start: %s", err.Error())
	}
}
