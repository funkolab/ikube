package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var version = "0.0.0"

func parseFlags() (string, appConfig) {
	var config appConfig
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  ikube [flags] [filter]\n\n")
		flag.PrintDefaults()
	}
	verbose := flag.Bool("v", false, "verbose mode")
	temp := flag.Bool("l", false, "load kubeconfig in temporary shell")
	delete := flag.Bool("d", false, "delete kubeconfig(s)")
	showVersion := flag.Bool("version", false, "display version")
	flag.Parse()

	config.verbose = *verbose
	config.temp = *temp
	config.delete = *delete

	// Check if version flag is set
	if *showVersion {
		fmt.Printf("ikube version %s\n", version)
		os.Exit(0)
	}

	// Get filter from remaining args
	var filter string
	if flag.NArg() > 0 {
		filter = strings.ToLower(flag.Arg(0))
	}

	return filter, config
}

func main() {
	// Parse command line flags
	filter, config := parseFlags()

	// Get project ID from environment variable
	projectID := os.Getenv("INFISICAL_PROJECT_ID")
	if projectID == "" {
		fmt.Println("Error: INFISICAL_PROJECT_ID environment variable is not set")
		os.Exit(1)
	}

	// Authenticate with Infisical
	client, err := authenticateInfisical(config)
	if err != nil {
		if config.verbose {
			fmt.Printf("Failed to authenticate: %v\n", err)
		} else {
			fmt.Println("Failed to authenticate")
		}
		os.Exit(1)
	}

	if config.delete {
		handleDeleteKubeconfigs(client, projectID, filter, config)
		return
	}

	// Check if we have stdin input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Has stdin input - store new kubeconfig
		handleStoreKubeconfig(client, projectID, config)
	} else {
		// No stdin input - list and select secrets
		handleListSecrets(client, projectID, filter, config)
	}
}
