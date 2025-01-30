package main

import (
	"fmt"
	"os"
	"os/exec"
)

func launchShellWithKubeconfig(kubeconfigPath string, config appConfig) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set up environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("KUBECONFIG=%s", kubeconfigPath))
	cmd.Env = env

	// Print information about the temporary session
	fmt.Printf("\nStarting temporary shell with KUBECONFIG=%s\n", kubeconfigPath)
	fmt.Println("Exit the shell to clean up the temporary kubeconfig")
	fmt.Println()

	// Run the shell
	err := cmd.Run()
	if err != nil {
		if config.verbose {
			return fmt.Errorf("error running shell: %v", err)
		}
		return fmt.Errorf("error running shell")
	}

	return nil
}

func createTempKubeconfig(content []byte) (string, error) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}

	// Write content
	if err := os.WriteFile(tmpfile.Name(), content, 0600); err != nil {
		os.Remove(tmpfile.Name())
		return "", fmt.Errorf("failed to write temporary file: %v", err)
	}

	return tmpfile.Name(), nil
}
