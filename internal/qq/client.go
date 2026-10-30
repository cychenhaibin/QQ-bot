package qq

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"
)

type ClientConfig struct {
	HTTPClient    *http.Client
	APIBaseURL    string
	AppID         string
	AppSecret     string
	TokenEndpoint string
}

type Client struct {
	httpClient    *http.Client
	apiBaseURL    string
	appID         string
	appSecret     string
	tokenEndpoint string

	mu                sync.Mutex
	cachedAccessToken string
	tokenExpiry       time.Time
}

func NewClient(cfg ClientConfig) *Client {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	apiBaseURL := strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = "https://api.bot.qq.com"
	}
	tokenEndpoint := strings.TrimSpace(cfg.TokenEndpoint)
	if tokenEndpoint == "" {
		tokenEndpoint = "https://bots.qq.com/app/getAppAccessToken"
	}
	return &Client{
		httpClient:    httpClient,
		apiBaseURL:    apiBaseURL,
		appID:         cfg.AppID,
		appSecret:     cfg.AppSecret,
		tokenEndpoint: tokenEndpoint,
	}
}

func (c *Client) SendGroupMessage(ctx context.Context, groupOpenID, content, messageID string) error {
	if strings.TrimSpace(groupOpenID) == "" {
		return errors.New("group openid is required")
	}
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}
	payload := map[string]string{"content": content}
	if messageID != "" {
		payload["msg_id"] = messageID
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	endpoint := c.apiBaseURL + "/v2/groups/" + url.PathEscape(groupOpenID) + "/messages"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "QQBot "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("send QQ group message: status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func (c *Client) AccessToken(ctx context.Context) (string, error) {
	return c.accessToken(ctx)
}

func (c *Client) GetGatewayURL(ctx context.Context) (string, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return "", err
	}
	endpoint := c.apiBaseURL + "/gateway"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "QQBot "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return "", fmt.Errorf("get QQ gateway URL: status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var response struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	if response.URL == "" {
		return "", errors.New("QQ gateway response is missing url")
	}
	return response.URL, nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	if c.cachedAccessToken != "" && time.Until(c.tokenExpiry) > time.Minute {
		token := c.cachedAccessToken
		c.mu.Unlock()
		return token, nil
	}
	c.mu.Unlock()

	payload, err := json.Marshal(map[string]string{"appId": c.appID, "clientSecret": c.appSecret})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return "", fmt.Errorf("get QQ access token: status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}
	var tokenResponse struct {
		AccessToken string          `json:"access_token"`
		ExpiresIn   json.RawMessage `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}
	expiresIn, err := parseExpiresIn(tokenResponse.ExpiresIn)
	if err != nil {
		return "", fmt.Errorf("parse QQ access token expiry: %w", err)
	}
	if tokenResponse.AccessToken == "" || expiresIn <= 0 {
		return "", errors.New("QQ access token response is invalid")
	}
	c.mu.Lock()
	c.cachedAccessToken = tokenResponse.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
	c.mu.Unlock()
	return tokenResponse.AccessToken, nil
}

func parseExpiresIn(raw json.RawMessage) (int64, error) {
	var numeric int64
	if err := json.Unmarshal(raw, &numeric); err == nil {
		return numeric, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(text), 10, 64)
}
