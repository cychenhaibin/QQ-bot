package newapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type AIRequest struct {
	MemberOpenID string `json:"member_openid"`
	Group        string `json:"group,omitempty"`
}

type AIGroup struct {
	Group  string   `json:"group"`
	Models []string `json:"models"`
}

type AIAccessResponse struct {
	APIKey string `json:"api_key"`
	Group  string `json:"group"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) ListAIGroups(ctx context.Context, request AIRequest) ([]AIGroup, error) {
	if request.MemberOpenID == "" {
		return nil, errors.New("member openid is required")
	}
	var response struct {
		Groups []AIGroup `json:"groups"`
	}
	if err := c.DoJSON(ctx, http.MethodPost, "/api/bot/v1/ai/groups", request, &response); err != nil {
		return nil, err
	}
	return response.Groups, nil
}

func (c *Client) GetAIAccess(ctx context.Context, request AIRequest) (AIAccessResponse, error) {
	if request.MemberOpenID == "" || request.Group == "" {
		return AIAccessResponse{}, errors.New("member openid and group are required")
	}
	var response AIAccessResponse
	if err := c.DoJSON(ctx, http.MethodPost, "/api/bot/v1/ai/access", request, &response); err != nil {
		return AIAccessResponse{}, err
	}
	if response.APIKey == "" {
		return AIAccessResponse{}, errors.New("new-api returned an empty AI API key")
	}
	return response, nil
}

func (c *Client) Chat(ctx context.Context, apiKey, model string, messages []ChatMessage) (string, error) {
	if apiKey == "" || model == "" || len(messages) == 0 {
		return "", errors.New("api key, model and messages are required")
	}
	body, err := json.Marshal(map[string]any{"model": model, "messages": messages, "stream": false})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("AI chat request failed: status %d", resp.StatusCode)
	}
	var result struct {
		Choices []struct {
			Message ChatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", errors.New("AI chat response did not contain a message")
	}
	return result.Choices[0].Message.Content, nil
}
