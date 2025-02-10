package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"connector-recruitment/go-server/connectors/logger"
)

const BaseUrl = "https://slack.com/api"

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseURL    string
	httpClient HttpClient
	logger     logger.Logger
}

func NewClient(baseURL string, httpClient HttpClient, logger logger.Logger) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

type SlackResponse struct {
	Ok      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
	Message struct {
		Text        string `json:"text,omitempty"`
		Username    string `json:"username,omitempty"`
		BotID       string `json:"bot_id,omitempty"`
		Attachments []struct {
			Text     string `json:"text,omitempty"`
			ID       int    `json:"id,omitempty"`
			Fallback string `json:"fallback,omitempty"`
		} `json:"attachments,omitempty"`
		Type    string `json:"type,omitempty"`
		Subtype string `json:"subtype,omitempty"`
		TS      string `json:"ts,omitempty"`
	} `json:"message,omitempty"`
}

func (c *Client) SendMessageToChannel(ctx context.Context, token, channelID, msg string) error {
	payload := map[string]string{
		"channel": channelID,
		"text":    msg,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat.postMessage", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token) // Add Authorization token

	// executes the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var slackResp SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// handle slack response
	if !slackResp.Ok {
		return fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	c.logger.Info("Message sent successfully", "slack-channel", slackResp.Channel, "timestamp", slackResp.TS)
	return nil
}
