package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	infisical "github.com/infisical/go-sdk"
	"github.com/zalando/go-keyring"
)

// clearStoredCredentials removes credentials from keyring
func clearStoredCredentials() {
	_ = keyring.Delete(keyringService, clientIDKey)
	_ = keyring.Delete(keyringService, clientSecretKey)
}

// storeCredentials saves valid credentials to keyring
func storeCredentials(clientID, clientSecret string) error {
	if err := keyring.Set(keyringService, clientIDKey, clientID); err != nil {
		return fmt.Errorf("failed to store client ID in keyring: %v", err)
	}

	if err := keyring.Set(keyringService, clientSecretKey, clientSecret); err != nil {
		// If we failed to store the secret, remove the ID as well
		_ = keyring.Delete(keyringService, clientIDKey)
		return fmt.Errorf("failed to store client secret in keyring: %v", err)
	}

	return nil
}

func getCredentials(forcePrompt bool) (string, string, bool, error) {
	if !forcePrompt {
		// First try environment variables
		clientID := os.Getenv("INFISICAL_CLIENT_ID")
		clientSecret := os.Getenv("INFISICAL_CLIENT_SECRET")

		// If both are set in env vars, use them
		if clientID != "" && clientSecret != "" {
			return clientID, clientSecret, false, nil
		}

		// Try to get from keyring
		var err error
		if clientID == "" {
			clientID, err = keyring.Get(keyringService, clientIDKey)
			if err != nil && err != keyring.ErrNotFound {
				return "", "", false, fmt.Errorf("failed to get client ID from keyring: %v", err)
			}
		}

		if clientSecret == "" {
			clientSecret, err = keyring.Get(keyringService, clientSecretKey)
			if err != nil && err != keyring.ErrNotFound {
				return "", "", false, fmt.Errorf("failed to get client secret from keyring: %v", err)
			}
		}

		// If we found both in keyring, return them
		if clientID != "" && clientSecret != "" {
			return clientID, clientSecret, true, nil
		}
	}

	// If we get here, we need to prompt for credentials
	return promptForCredentials()
}

func promptForCredentials() (string, string, bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Infisical Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return "", "", false, fmt.Errorf("failed to read client ID: %v", err)
	}
	clientID = strings.TrimSpace(clientID)

	fmt.Print("Enter Infisical Client Secret: ")
	clientSecret, err := reader.ReadString('\n')
	if err != nil {
		return "", "", false, fmt.Errorf("failed to read client secret: %v", err)
	}
	clientSecret = strings.TrimSpace(clientSecret)

	return clientID, clientSecret, false, nil
}

func authenticateInfisical(config appConfig) (infisical.InfisicalClientInterface, error) {
	// First attempt with stored or env credentials
	clientID, clientSecret, isFromKeyring, err := getCredentials(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %v", err)
	}

	client := infisical.NewInfisicalClient(context.Background(), infisical.Config{
		SiteUrl:          fmt.Sprintf("https://%s", config.infisicalServer),
		AutoTokenRefresh: true,
	})

	_, err = client.Auth().UniversalAuthLogin(clientID, clientSecret)
	if err == nil {
		// If credentials were manually entered and valid, store them
		if !isFromKeyring {
			if err := storeCredentials(clientID, clientSecret); err != nil {
				if config.verbose {
					fmt.Printf("Warning: Failed to store credentials: %v\n", err)
				} else {
					fmt.Println("Warning: Failed to store credentials")
				}
			}
		}
		return client, nil
	}

	// If credentials were from keyring and invalid, clear them and try once more
	if isFromKeyring {
		if config.verbose {
			fmt.Printf("Stored credentials are invalid: %v\n", err)
		} else {
			fmt.Println("Stored credentials are invalid")
		}
		clearStoredCredentials()

		// Second attempt with manual input
		clientID, clientSecret, _, err = getCredentials(true)
		if err != nil {
			return nil, fmt.Errorf("failed to get credentials: %v", err)
		}

		client = infisical.NewInfisicalClient(context.Background(), infisical.Config{
			SiteUrl:          fmt.Sprintf("https://%s", config.infisicalServer),
			AutoTokenRefresh: true,
		})

		_, err = client.Auth().UniversalAuthLogin(clientID, clientSecret)
		if err == nil {
			// Store the valid credentials
			if err := storeCredentials(clientID, clientSecret); err != nil {
				if config.verbose {
					fmt.Printf("Warning: Failed to store credentials: %v\n", err)
				} else {
					fmt.Println("Warning: Failed to store credentials")
				}
			}
			return client, nil
		}

		if config.verbose {
			return nil, fmt.Errorf("authentication failed with new credentials: %v", err)
		}
		return nil, fmt.Errorf("authentication failed with new credentials")
	}

	// If credentials were from env vars or manual input and failed, exit
	if config.verbose {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}
	return nil, fmt.Errorf("authentication failed")
}
