package handler

import (
	"context"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

const RechargeShopURL = "https://catfk.com/shop/FluxA"

type RechargeHandler struct {
	ShopURL    string
	configured bool
}

func NewRechargeHandler(shopURL string) RechargeHandler {
	return RechargeHandler{ShopURL: shopURL, configured: true}
}

func (h RechargeHandler) Handle(_ context.Context, event qq.GroupMessage) (string, error) {
	if event.Content != "充值中心" {
		return "", nil
	}
	shopURL := h.ShopURL
	if h.configured && shopURL == "" {
		return "充值链接暂未配置，请联系管理员。\n⚠️付费站和公益站是两个不同的站子，请看清楚再下单", nil
	}
	if shopURL == "" {
		shopURL = RechargeShopURL
	}
	return "充值链接：" + shopURL + "\n⚠️付费站和公益站是两个不同的站子，请看清楚再下单", nil
}
