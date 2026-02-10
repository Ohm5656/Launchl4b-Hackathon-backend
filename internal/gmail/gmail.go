package gmail

import (
	"context"
	"log"
	"sync"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"net/http"
)

type MessageInfo struct {
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
}

type FetchResult struct {
	TotalFound  int           `json:"total_found"`
	FetchedCount int           `json:"fetched_count"`
	Messages    []MessageInfo `json:"messages"`
	ErrorsCount int           `json:"errors_count"`
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

			msg, err := srv.Users.Messages.Get(user, id).Format("metadata").MetadataHeaders("From", "Subject").Do()
			if err != nil {
				log.Printf("Error fetching %s: %v", id, err)
				errChan <- err
				return
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
