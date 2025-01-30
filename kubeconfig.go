package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	infisical "github.com/infisical/go-sdk"
	"github.com/ktr0731/go-fuzzyfinder"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func validateKubeconfig(config *api.Config) error {
	if config == nil {
		return fmt.Errorf("kubeconfig is nil")
	}

	// Check if there's at least one context
	if len(config.Contexts) == 0 {
		return fmt.Errorf("no contexts found in kubeconfig")
	}

	// Check if there's at least one cluster
	if len(config.Clusters) == 0 {
		return fmt.Errorf("no clusters found in kubeconfig")
	}

	// Validate current context exists
	if config.CurrentContext == "" {
		return fmt.Errorf("no current-context set in kubeconfig")
	}

	// Get current context and validate it exists in contexts
	context, exists := config.Contexts[config.CurrentContext]
	if !exists {
		return fmt.Errorf("current-context '%s' not found in contexts", config.CurrentContext)
	}

	// Get cluster name and validate it exists
	clusterName := context.Cluster
	cluster, exists := config.Clusters[clusterName]
	if !exists {
		return fmt.Errorf("cluster '%s' referenced by context '%s' not found in clusters", 
			clusterName, config.CurrentContext)
	}

	// Validate cluster server URL
	if cluster.Server == "" {
		return fmt.Errorf("cluster '%s' has no server URL specified", clusterName)
	}

	// Validate auth info
	if context.AuthInfo == "" {
		return fmt.Errorf("no user/auth info specified in context '%s'", config.CurrentContext)
	}

	authInfo, exists := config.AuthInfos[context.AuthInfo]
	if !exists {
		return fmt.Errorf("user '%s' referenced by context '%s' not found in users", 
			context.AuthInfo, config.CurrentContext)
	}

	// Check if at least one authentication method is specified
	hasAuth := authInfo.Token != "" || 
		authInfo.ClientCertificate != "" || 
		authInfo.ClientCertificateData != nil ||
		authInfo.ClientKey != "" ||
		authInfo.ClientKeyData != nil ||
		authInfo.TokenFile != "" ||
		authInfo.Exec != nil

	if !hasAuth {
		return fmt.Errorf("no authentication method specified for user '%s'", context.AuthInfo)
	}

	return nil
}

func handleStoreKubeconfig(client infisical.InfisicalClientInterface, projectID string, config appConfig) {
	// Read kubeconfig from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var kubeconfig string
	for scanner.Scan() {
		kubeconfig += scanner.Text() + "\n"
	}

	if err := scanner.Err(); err != nil {
		if config.verbose {
			fmt.Printf("Error reading from stdin: %v\n", err)
		} else {
			fmt.Println("Error reading from stdin")
		}
		os.Exit(1)
	}

	// Check if input is empty
	if strings.TrimSpace(kubeconfig) == "" {
		fmt.Println("Error: Empty kubeconfig received")
		os.Exit(1)
	}

	// Parse kubeconfig
	kubeCfg, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		if config.verbose {
			fmt.Printf("Error: Invalid kubeconfig format: %v\n", err)
		} else {
			fmt.Println("Error: Invalid kubeconfig format")
		}
		os.Exit(1)
	}

	// Validate kubeconfig structure
	if err := validateKubeconfig(kubeCfg); err != nil {
		if config.verbose {
			fmt.Printf("Error: Invalid kubeconfig: %v\n", err)
		} else {
			fmt.Println("Error: Invalid kubeconfig")
		}
		os.Exit(1)
	}

	// Get cluster details
	currentContext := kubeCfg.CurrentContext
	clusterName := kubeCfg.Contexts[currentContext].Cluster
	serverAddress := kubeCfg.Clusters[clusterName].Server

	// First, check if the secret already exists
	secrets, err := client.Secrets().List(infisical.ListSecretsOptions{
		ProjectID:          projectID,
		Environment:        "config",
		SecretPath:         "/",
		AttachToProcessEnv: false,
	})
	if err != nil {
		if config.verbose {
			fmt.Printf("Failed to check existing secrets: %v\n", err)
		} else {
			fmt.Println("Failed to check existing secrets")
		}
		os.Exit(1)
	}

	var existingSecret *infisical.Secret
	for _, secret := range secrets {
		if secret.SecretKey == clusterName {
			existingSecret = &secret
			break
		}
	}

	if existingSecret != nil {
		// Update existing secret
		_, err = client.Secrets().Update(infisical.UpdateSecretOptions{
			ProjectID:      projectID,
			Environment:    "config",
			SecretKey:      clusterName,
			NewSecretValue: kubeconfig,
		})
		if err != nil {
			if config.verbose {
				fmt.Printf("Failed to update secret: %v\n", err)
			} else {
				fmt.Println("Failed to update secret")
			}
			os.Exit(1)
		}
		fmt.Printf("Successfully updated kubeconfig for cluster: %s\n", clusterName)
	} else {
		// Create new secret
		_, err = client.Secrets().Create(infisical.CreateSecretOptions{
			ProjectID:     projectID,
			Environment:   "config",
			SecretKey:     clusterName,
			SecretValue:   kubeconfig,
			SecretComment: fmt.Sprintf("Cluster: %s\nServer: %s", clusterName, serverAddress),
		})
		if err != nil {
			if config.verbose {
				fmt.Printf("Failed to store secret: %v\n", err)
			} else {
				fmt.Println("Failed to store secret")
			}
			os.Exit(1)
		}
		fmt.Printf("Successfully stored kubeconfig for cluster: %s\n", clusterName)
	}
}

func handleListSecrets(client infisical.InfisicalClientInterface, projectID string, filter string, config appConfig) {
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

	var selectedSecret infisical.Secret
	if len(secrets) == 1 {
		// If there's only one result, use it directly
		selectedSecret = secrets[0]
		fmt.Printf("Using only available kubeconfig: %s\n", selectedSecret.SecretKey)
	} else {
		// Use fuzzy finder to select a kubeconfig
		idx, err := fuzzyfinder.Find(
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

		selectedSecret = secrets[idx]
	}

	if config.temp {
		// Create temporary kubeconfig file
		tmpPath, err := createTempKubeconfig([]byte(selectedSecret.SecretValue))
		if err != nil {
			if config.verbose {
				fmt.Printf("Error creating temporary kubeconfig: %v\n", err)
			} else {
				fmt.Println("Error creating temporary kubeconfig")
			}
			os.Exit(1)
		}
		
		// Ensure cleanup of temporary file
		defer os.Remove(tmpPath)

		// Launch shell with temporary kubeconfig
		err = launchShellWithKubeconfig(tmpPath, config)
		if err != nil {
			if config.verbose {
				fmt.Printf("Error launching shell: %v\n", err)
			} else {
				fmt.Println("Error launching shell")
			}
			os.Exit(1)
		}
		return
	}

	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		if config.verbose {
			fmt.Printf("Error getting home directory: %v\n", err)
		} else {
			fmt.Println("Error getting home directory")
		}
		os.Exit(1)
	}

	// Create .kube directory if it doesn't exist
	kubeDir := filepath.Join(homeDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		if config.verbose {
			fmt.Printf("Error creating .kube directory: %v\n", err)
		} else {
			fmt.Println("Error creating .kube directory")
		}
		os.Exit(1)
	}

	// Write kubeconfig to file
	kubeconfigPath := filepath.Join(kubeDir, "config")
	if err := os.WriteFile(kubeconfigPath, []byte(selectedSecret.SecretValue), 0600); err != nil {
		if config.verbose {
			fmt.Printf("Error writing kubeconfig: %v\n", err)
		} else {
			fmt.Println("Error writing kubeconfig")
		}
		os.Exit(1)
	}

	fmt.Printf("Successfully configured kubeconfig for cluster: %s\n", selectedSecret.SecretKey)
}
