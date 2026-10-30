package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

func TestRechargeHandlerReturnsConfiguredShopAndWarning(t *testing.T) {
	reply, err := (RechargeHandler{}).Handle(context.Background(), qq.GroupMessage{Content: "充值中心"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, RechargeShopURL) || !strings.Contains(reply, "⚠️付费站和公益站是两个不同的站子，请看清楚再下单") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestRechargeHandlerIgnoresOtherMessages(t *testing.T) {
	reply, err := (RechargeHandler{}).Handle(context.Background(), qq.GroupMessage{Content: "查询余额"})
	if err != nil || reply != "" {
		t.Fatalf("Handle() = %q, %v", reply, err)
	}
}

func TestRechargeHandlerDoesNotUseAnotherSitesURLWhenUnset(t *testing.T) {
	reply, err := NewRechargeHandler("").Handle(context.Background(), qq.GroupMessage{Content: "充值中心"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(reply, RechargeShopURL) || !strings.Contains(reply, "充值链接暂未配置") {
		t.Fatalf("unexpected unset-site reply: %s", reply)
	}
}
