package main

import (
	"fmt"
	"log"
	"net/http"

	"gmail-fetcher-web/internal/auth"
	"gmail-fetcher-web/internal/handlers"
)

func main() {
	// Initialize OAuth Config
	err := auth.LoadConfig("credentials.json")
	if err != nil {
		log.Printf("Warning: Could not load credentials.json: %v", err)
	}

	// Routes
	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/login", handlers.HandleLogin)
	http.HandleFunc("/callback", handlers.HandleCallback)

	port := ":8080"
	fmt.Printf("Gmail Fetcher Web started at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
