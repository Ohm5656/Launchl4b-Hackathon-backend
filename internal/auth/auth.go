package auth

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var Config *oauth2.Config

func LoadConfig(credentialsPath string) error {
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return err
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return err
	}
	// Default redirect URL for local dev
	config.RedirectURL = "http://localhost:8080/callback"
	Config = config
	return nil
}
