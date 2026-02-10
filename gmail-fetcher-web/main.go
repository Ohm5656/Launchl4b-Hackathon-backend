package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// MessageInfo stores email details
type MessageInfo struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
}

var (
	googleOauthConfig *oauth2.Config
	// In production, use a secure way to store states and sessions.
	// This is a simplified version for demonstration.
)

func init() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Printf("Warning: Unable to read credentials.json: %v", err)
		return
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	// Ensure RedirectURL is set to our callback
	config.RedirectURL = "http://localhost:8080/callback"
	googleOauthConfig = config
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if googleOauthConfig == nil {
		http.Error(w, "Google Credentials not configured. Please check credentials.json", http.StatusInternalServerError)
		return
	}
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "No code found", http.StatusBadRequest)
		return
	}

	tok, err := googleOauthConfig.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := googleOauthConfig.Client(context.TODO(), tok)
	srv, err := gmail.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		http.Error(w, "Failed to create Gmail service", http.StatusInternalServerError)
		return
	}

	user := "me"
	res, err := srv.Users.Messages.List(user).MaxResults(100).Do()
	if err != nil {
		http.Error(w, "Failed to list messages", http.StatusInternalServerError)
		return
	}

	var messages []MessageInfo
	for _, m := range res.Messages {
		msg, err := srv.Users.Messages.Get(user, m.Id).Format("metadata").MetadataHeaders("From", "Subject").Do()
		if err != nil {
			continue
		}

		info := MessageInfo{Snippet: msg.Snippet}
		for _, h := range msg.Payload.Headers {
			if h.Name == "From" {
				info.From = h.Value
			}
			if h.Name == "Subject" {
				info.Subject = h.Value
			}
		}
		messages = append(messages, info)
	}

	// For simplicity, we just display the results as JSON on the callback page.
	// In a real app, we'd render another HTML template.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)

	// Proactively save to file as well, as requested
	saveBytes, _ := json.MarshalIndent(messages, "", "  ")
	os.WriteFile("inbox.json", saveBytes, 0644)
	log.Println("Inbox saved to inbox.json")
}

