package main

import (
	"log"
	"os"

	"github.com/hashicorp/vault/api"
)

func getJWTSecret() (string, error) {
	config := api.Config{
		Address: "http://127.0.0.1:8200",
	}
	client, err := api.NewClient(&config)
	if err != nil {
		return "", err
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))

	l := client.Logical()
	secret, err := l.Read("secret/hashiapp")
	if err != nil {
		return "", err
	}
	return secret.Data["jwtsecret"].(string), nil
}

func getDatabaseCredentials() (string, string, error) {
	config := api.Config{
		Address: "http://127.0.0.1:8200",
	}
	client, err := api.NewClient(&config)
	if err != nil {
		return "", "", err
	}
	client.SetToken(os.Getenv("VAULT_TOKEN"))

	l := client.Logical()
	secret, err := l.Read("mysql/creds/hashiapp")
	if err != nil {
		return "", "", err
	}
	username := secret.Data["username"].(string)
	password := secret.Data["password"].(string)
	log.Println("Username:", username)
	log.Println("Password:", password)
	return username, password, nil
}
