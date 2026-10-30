package newapi

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ClientConfig struct {
	HTTPClient *http.Client
	BaseURL    string
	BotID      string
	Secret     string
	Now        func() time.Time
	Nonce      func() (string, error)
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	botID      string
	secret     string
	now        func() time.Time
	nonce      func() (string, error)
}

func NewClient(cfg ClientConfig) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	nonce := cfg.Nonce
	if nonce == nil {
		nonce = randomNonce
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/"),
		botID:      cfg.BotID,
		secret:     cfg.Secret,
		now:        now,
		nonce:      nonce,
	}
}

func (c *Client) DoJSON(ctx context.Context, method, path string, requestBody, responseBody any) error {
	if c.baseURL == "" {
		return errors.New("new-api base URL is required")
	}
	if c.botID == "" || c.secret == "" {
		return errors.New("new-api Bot credentials are required")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	body, err := json.Marshal(requestBody)
	if requestBody == nil {
		body = nil
	}
	if err != nil {
		return err
	}
	timestamp := c.now().Unix()
	nonce, err := c.nonce()
	if err != nil {
		return fmt.Errorf("generate Bot request nonce: %w", err)
	}
	canonical := strings.ToUpper(method) + "\n" + path + "\n" + fmt.Sprint(timestamp) + "\n" + nonce + "\n" + string(body)
	mac := hmac.New(sha256.New, []byte(c.secret))
	_, _ = mac.Write([]byte(canonical))
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-Bot-ID", c.botID)
	req.Header.Set("X-Bot-Timestamp", fmt.Sprint(timestamp))
	req.Header.Set("X-Bot-Nonce", nonce)
	req.Header.Set("X-Bot-Signature", signature)
	req.Header.Set("X-Request-ID", nonce)
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("new-api request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBodyBytes)))
	}
	if responseBody == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	responseBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read new-api response: %w", err)
	}
	var envelope struct {
		Success *bool           `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(responseBytes, &envelope); err == nil && envelope.Success != nil {
		if !*envelope.Success {
			if envelope.Message == "" {
				envelope.Message = "new-api request failed"
			}
			return errors.New(envelope.Message)
		}
		if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
			if err := json.Unmarshal(envelope.Data, responseBody); err != nil {
				return fmt.Errorf("decode new-api response data: %w", err)
			}
			return nil
		}
	}
	if err := json.Unmarshal(responseBytes, responseBody); err != nil {
		return fmt.Errorf("decode new-api response: %w", err)
	}
	return nil
}

func randomNonce() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
