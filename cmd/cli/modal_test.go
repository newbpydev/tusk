package main

import (
	"flag"
	"os"
	
	"github.com/newbpydev/tusk/internal/examples/modal"
)

// This file contains a command line flag for testing the modal example
// Usage: go run cmd/cli/main.go -modal
// This allows testing the modal functionality without modifying the main application

var runModalExample = flag.Bool("modal", false, "Run the modal example instead of the main application")

func init() {
	// We need to parse flags early to check for the modal flag
	flag.Parse()
	
	// If the modal flag is set, run the modal example and exit
	if *runModalExample {
		modal.RunModalExample()
		// Exit after running the example to prevent the main app from starting
		os.Exit(0)
	}
}
