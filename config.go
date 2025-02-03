package main

type appConfig struct {
	verbose         bool
	temp            bool
	delete          bool
	infisicalServer string
}

const (
	keyringService  = "kube-infisical"
	clientIDKey     = "client_id"
	clientSecretKey = "client_secret"
)
