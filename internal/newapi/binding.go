package newapi

import (
	"context"
	"errors"
)

type StartBindingRequest struct {
	MemberOpenID string `json:"member_openid"`
	Email        string `json:"email"`
	ForceReplace bool   `json:"force_replace,omitempty"`
}

type StartBindingResponse struct {
	FlowToken       string `json:"flow_token"`
	RequiresReplace bool   `json:"requires_replace"`
	ExistingDomain  string `json:"existing_domain,omitempty"`
}

type VerifyBindingRequest struct {
	MemberOpenID string `json:"member_openid"`
	FlowToken    string `json:"flow_token"`
	Code         string `json:"code"`
}

type VerifyBindingResponse struct {
	Username string `json:"username"`
}

func (c *Client) StartBinding(ctx context.Context, request StartBindingRequest) (StartBindingResponse, error) {
	if request.MemberOpenID == "" || request.Email == "" {
		return StartBindingResponse{}, errors.New("member openid and email are required")
	}
	var response StartBindingResponse
	if err := c.DoJSON(ctx, "POST", "/api/bot/v1/bindings/start", request, &response); err != nil {
		return StartBindingResponse{}, err
	}
	if !request.ForceReplace && response.FlowToken == "" && !response.RequiresReplace {
		return StartBindingResponse{}, errors.New("new-api binding response is invalid")
	}
	return response, nil
}

func (c *Client) VerifyBinding(ctx context.Context, request VerifyBindingRequest) (VerifyBindingResponse, error) {
	if request.MemberOpenID == "" || request.FlowToken == "" || request.Code == "" {
		return VerifyBindingResponse{}, errors.New("member openid, flow token and code are required")
	}
	var response VerifyBindingResponse
	if err := c.DoJSON(ctx, "POST", "/api/bot/v1/bindings/verify", request, &response); err != nil {
		return VerifyBindingResponse{}, err
	}
	return response, nil
}
