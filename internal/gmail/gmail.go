package gmail

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type MessageInfo struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Snippet string `json:"snippet"`
	Detail  string `json:"detail"`
	Price   string `json:"price,omitempty"`
}

type FetchResult struct {
	TotalFound   int           `json:"total_found"`
	FetchedCount int           `json:"fetched_count"`
	Messages     []MessageInfo `json:"messages"`
	ErrorsCount  int           `json:"errors_count"`
}

func FetchInbox(ctx context.Context, client *http.Client, maxResults int64) (*FetchResult, error) {
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	user := "me"
	res, err := srv.Users.Messages.List(user).MaxResults(maxResults).Do()
	if err != nil {
		return nil, err
	}

	totalMessages := len(res.Messages)
	messages := make([]MessageInfo, 0, totalMessages)
	msgChan := make(chan MessageInfo, totalMessages)
	errChan := make(chan error, totalMessages)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for _, m := range res.Messages {
		wg.Add(1)
		sem <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Use 'full' to get the body parts
			msg, err := srv.Users.Messages.Get(user, id).Format("full").Do()
			if err != nil {
				log.Printf("Error fetching %s: %v", id, err)
				errChan <- err
				return
			}

			detail := getBodyContent(msg.Payload)
			info := MessageInfo{
				Snippet: msg.Snippet,
				Detail:  detail,
			}

			// Extract price from snippet or detail
			info.Price = extractPrice(info.Subject + " " + info.Snippet + " " + detail)

			for _, h := range msg.Payload.Headers {
				switch h.Name {
				case "From":
					info.From = h.Value
				case "Subject":
					info.Subject = h.Value
				case "Date":
					info.Date = h.Value // Keeping the original full date string as requested
				}
			}
			msgChan <- info
		}(m.Id)
	}

	go func() {
		wg.Wait()
		close(msgChan)
		close(errChan)
	}()

	for m := range msgChan {
		messages = append(messages, m)
	}

	return &FetchResult{
		TotalFound:   totalMessages,
		FetchedCount: len(messages),
		Messages:     messages,
		ErrorsCount:  len(errChan),
	}, nil
}

// getBodyContent extracts the body and strips HTML if necessary
func getBodyContent(payload *gmail.MessagePart) string {
	// 1. Try to find text/plain first
	plain := findPart(payload, "text/plain")
	if plain != "" {
		return plain
	}

	// 2. If no plain text, take text/html and strip tags
	html := findPart(payload, "text/html")
	if html != "" {
		return stripHTML(html)
	}

	return ""
}

func findPart(payload *gmail.MessagePart, mimeType string) string {
	if payload.MimeType == mimeType && payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	for _, part := range payload.Parts {
		if content := findPart(part, mimeType); content != "" {
			return content
		}
	}
	return ""
}

func stripHTML(html string) string {
	// Basic regex to strip HTML tags
	re := regexp.MustCompile("<[^>]*>")
	plain := re.ReplaceAllString(html, " ")
	
	// Clean up multiple spaces and entities
	plain = strings.ReplaceAll(plain, "&nbsp;", " ")
	plain = strings.ReplaceAll(plain, "&amp;", "&")
	plain = strings.ReplaceAll(plain, "&lt;", "<")
	plain = strings.ReplaceAll(plain, "&gt;", ">")
	plain = strings.ReplaceAll(plain, "&quot;", "\"")
	
	// Collapse multiple spaces/newlines
	spaceRe := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(spaceRe.ReplaceAllString(plain, " "))
}

// extractPrice looks for currency markers (฿, $, THB) and associated numbers
func extractPrice(content string) string {
	// Pattern 1: Symbol before number (e.g., ฿1,200, $50)
	// Supports comma as thousands separator and dot for decimals
	re1 := regexp.MustCompile(`([฿\$]|THB|USD)\s?(\d{1,3}(,\d{3})*(\.\d{1,2})?)`)
	match := re1.FindStringSubmatch(content)
	if len(match) > 0 {
		return match[1] + match[2]
	}

	// Pattern 2: Number before symbol (e.g., 500 บาท, 200 THB)
	re2 := regexp.MustCompile(`(\d{1,3}(,\d{3})*(\.\d{1,2})?)\s?(บาท|THB|USD|฿|\$)`)
	match = re2.FindStringSubmatch(content)
	if len(match) > 0 {
		return match[1] + " " + match[4]
	}

	return ""
}
