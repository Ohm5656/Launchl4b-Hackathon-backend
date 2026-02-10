package handlers

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"os"

	"gmail-fetcher-web/internal/auth"
	"gmail-fetcher-web/internal/gmail"
	"golang.org/x/oauth2"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if auth.Config == nil {
		http.Error(w, "OAuth config not initialized", http.StatusInternalServerError)
		return
	}
	url := auth.Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Code missing", http.StatusBadRequest)
		return
	}

	tok, err := auth.Config.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	client := auth.Config.Client(context.TODO(), tok)
	result, err := gmail.FetchInbox(context.TODO(), client, 100)
	if err != nil {
		http.Error(w, "Fetch failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to JSON file
	saveData, _ := json.MarshalIndent(result.Messages, "", "  ")
	os.WriteFile("output/inbox.json", saveData, 0644)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
