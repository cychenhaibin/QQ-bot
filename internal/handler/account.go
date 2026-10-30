package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type AccountAPI interface {
	LookupAccount(context.Context, newapi.AccountRequest) (newapi.AccountResponse, error)
	CheckIn(context.Context, newapi.AccountRequest) (newapi.CheckinResponse, error)
}

type AccountHandler struct {
	api AccountAPI
}

func NewAccountHandler(api AccountAPI) *AccountHandler {
	return &AccountHandler{api: api}
}

func (h *AccountHandler) Handle(ctx context.Context, event qq.GroupMessage) (string, error) {
	if h.api == nil {
		return "", errors.New("account API is not configured")
	}
	switch event.Content {
	case "查询余额":
		result, err := h.api.LookupAccount(ctx, newapi.AccountRequest{MemberOpenID: event.MemberOpenID})
		if err != nil {
			return "查询余额失败，请先绑定账号或稍后重试。", err
		}
		quota := result.QuotaDisplay
		if quota == "" {
			quota = formatQuotaFallback(result.Quota)
		}
		return "用户名：" + result.Username + "\n余额：" + quota, nil
	case "每日签到":
		_, err := h.api.CheckIn(ctx, newapi.AccountRequest{MemberOpenID: event.MemberOpenID})
		if err != nil {
			if errors.Is(err, newapi.ErrCheckInDisabled) {
				return "当前站点未开启签到功能，暂时无法签到。", nil
			}
			return "签到失败，请先绑定账号或稍后重试。", err
		}
		return "✅ 签到成功，今日奖励已到账。", nil
	default:
		return "", nil
	}
}

func formatQuotaFallback(quota int) string {
	return strconv.Itoa(quota)
}
