package main

import (
	"fmt"
	"os"
	"strings"

	infisical "github.com/infisical/go-sdk"
	"github.com/ktr0731/go-fuzzyfinder"
	"k8s.io/client-go/tools/clientcmd"
)

func handleDeleteKubeconfigs(client infisical.InfisicalClientInterface, projectID string, filter string, config appConfig) {
	// Get all secrets
	secrets, err := client.Secrets().List(infisical.ListSecretsOptions{
		ProjectID:          projectID,
		Environment:        "config",
		SecretPath:         "/",
		AttachToProcessEnv: false,
	})
	if err != nil {
		if config.verbose {
			fmt.Printf("Failed to retrieve secrets: %v\n", err)
		} else {
			fmt.Println("Failed to retrieve secrets")
		}
		os.Exit(1)
	}

	if len(secrets) == 0 {
		fmt.Println("No kubeconfigs found")
		os.Exit(0)
	}

	// Filter secrets if a filter is provided
	if filter != "" {
		filteredSecrets := make([]infisical.Secret, 0)
		for _, secret := range secrets {
			if strings.Contains(strings.ToLower(secret.SecretKey), filter) {
				filteredSecrets = append(filteredSecrets, secret)
			}
		}
		secrets = filteredSecrets

		if len(secrets) == 0 {
			fmt.Printf("No kubeconfigs found matching filter: %s\n", filter)
			os.Exit(0)
		}
	}

	// Use fuzzy finder to select kubeconfigs to delete
	indices, err := fuzzyfinder.FindMulti(
		secrets,
		func(i int) string {
			return secrets[i].SecretKey
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			// Parse the kubeconfig to get cluster details
			kubeCfg, err := clientcmd.Load([]byte(secrets[i].SecretValue))
			if err != nil {
				return fmt.Sprintf("Error parsing kubeconfig: %v", err)
			}

			// Get cluster details
			clusterName := secrets[i].SecretKey
			cluster := kubeCfg.Clusters[clusterName]

			return fmt.Sprintf("Cluster: %s\nServer: %s\nComment: %s",
				clusterName,
				cluster.Server,
				secrets[i].SecretComment)
		}),
	)

	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			fmt.Println("Selection cancelled")
			return
		}
		if config.verbose {
			fmt.Printf("Error during selection: %v\n", err)
		} else {
			fmt.Println("Error during selection")
		}
		os.Exit(1)
	}

	if len(indices) == 0 {
		fmt.Println("No kubeconfigs selected for deletion")
		return
	}

	// Confirm deletion
	fmt.Println("\nSelected kubeconfigs for deletion:")
	for _, idx := range indices {
		fmt.Printf("- %s\n", secrets[idx].SecretKey)
	}
	fmt.Print("\nAre you sure you want to delete these kubeconfigs? [y/N]: ")
	
	var confirmation string
	fmt.Scanln(&confirmation)
	if strings.ToLower(confirmation) != "y" {
		fmt.Println("Deletion cancelled")
		return
	}

	// Delete selected kubeconfigs
	for _, idx := range indices {
		secret := secrets[idx]
		_, err := client.Secrets().Delete(infisical.DeleteSecretOptions{
			ProjectID:   projectID,
			Environment: "config",
			SecretKey:   secret.SecretKey,
		})
		if err != nil {
			if config.verbose {
				fmt.Printf("Failed to delete kubeconfig %s: %v\n", secret.SecretKey, err)
			} else {
				fmt.Printf("Failed to delete kubeconfig %s\n", secret.SecretKey)
			}
			continue
		}
		fmt.Printf("Successfully deleted kubeconfig: %s\n", secret.SecretKey)
	}
}
