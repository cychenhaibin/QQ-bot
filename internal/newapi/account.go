package newapi

import (
	"context"
	"errors"
	"strings"
)

var ErrCheckInDisabled = errors.New("check-in feature is not enabled")

type AccountRequest struct {
	MemberOpenID string `json:"member_openid"`
}

type AccountResponse struct {
	Username     string `json:"username"`
	Quota        int    `json:"quota"`
	QuotaDisplay string `json:"quota_display"`
}

type CheckinResponse struct {
	Username     string `json:"username"`
	QuotaAwarded int    `json:"quota_awarded"`
	CheckinDate  string `json:"checkin_date"`
}

func (c *Client) LookupAccount(ctx context.Context, request AccountRequest) (AccountResponse, error) {
	if request.MemberOpenID == "" {
		return AccountResponse{}, errors.New("member openid is required")
	}
	var response AccountResponse
	if err := c.DoJSON(ctx, "POST", "/api/bot/v1/account/lookup", request, &response); err != nil {
		return AccountResponse{}, err
	}
	return response, nil
}

func (c *Client) CheckIn(ctx context.Context, request AccountRequest) (CheckinResponse, error) {
	if request.MemberOpenID == "" {
		return CheckinResponse{}, errors.New("member openid is required")
	}
	var response CheckinResponse
	if err := c.DoJSON(ctx, "POST", "/api/bot/v1/checkin", request, &response); err != nil {
		if strings.EqualFold(strings.TrimSpace(err.Error()), ErrCheckInDisabled.Error()) {
			return CheckinResponse{}, ErrCheckInDisabled
		}
		return CheckinResponse{}, err
	}
	return response, nil
}
