package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// MessageInfo stores the retrieved email details
type MessageInfo struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Use a local server to capture the code
	codeCh := make(chan string)
	server := &http.Server{Addr: ":8080"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code != "" {
			fmt.Fprintf(w, "Authorization successful! You can close this window now.")
			codeCh <- code
		} else {
			fmt.Fprintf(w, "Authorization failed. No code found.")
		}
	})

	// Set RedirectURL to localhost
	config.RedirectURL = "http://localhost:8080"
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open, visit this link: \n%v\n", authURL)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Local server failed: %v", err)
		}
	}()

	authCode := <-codeCh
	server.Shutdown(context.Background())

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v. Please make sure credentials.json exists.", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := srv.Users.Messages.List(user).MaxResults(20).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	var messages []MessageInfo
	if len(r.Messages) == 0 {
		fmt.Println("No messages found.")
	} else {
		fmt.Println("Messages:")
		for _, m := range r.Messages {
			msg, err := srv.Users.Messages.Get(user, m.Id).Format("metadata").MetadataHeaders("From", "Subject").Do()
			if err != nil {
				log.Printf("Unable to retrieve message %v: %v", m.Id, err)
				continue
			}

			info := MessageInfo{
				Snippet: msg.Snippet,
			}

			for _, h := range msg.Payload.Headers {
				if h.Name == "From" {
					info.From = h.Value
				}
				if h.Name == "Subject" {
					info.Subject = h.Value
				}
			}
			messages = append(messages, info)
			fmt.Printf("- %s\n", info.Subject)
		}
	}

	// Save to JSON
	file, _ := json.MarshalIndent(messages, "", "  ")
	err = os.WriteFile("inbox.json", file, 0644)
	if err != nil {
		log.Fatalf("Unable to save inbox.json: %v", err)
	}

	fmt.Println("\nSuccessfully saved 20 messages to inbox.json")
}
