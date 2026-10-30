package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type fakeChannelAPI struct {
	groups []newapi.ChannelGroup
}

func (f fakeChannelAPI) ListChannels(context.Context) ([]newapi.ChannelGroup, error) {
	return f.groups, nil
}

func TestChannelHandlerListsGroupsAndModels(t *testing.T) {
	availability := 0.9327
	h := NewChannelHandler(fakeChannelAPI{groups: []newapi.ChannelGroup{{Group: "default", Availability7d: &availability}}})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{Content: "模型渠道"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "default") || !strings.Contains(reply, "7 日可用率：93.27%") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestChannelHandlerDistinguishesNoDataAndZeroAvailability(t *testing.T) {
	zero := 0.0
	h := NewChannelHandler(fakeChannelAPI{groups: []newapi.ChannelGroup{{Group: "empty"}, {Group: "failed", Availability7d: &zero}}})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{Content: "模型渠道"})
	if err != nil || !strings.Contains(reply, "7 日可用率：暂无数据") || !strings.Contains(reply, "7 日可用率：0.00%") {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestChannelHandlerIgnoresOtherMessages(t *testing.T) {
	h := NewChannelHandler(fakeChannelAPI{})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{Content: "充值中心"})
	if err != nil || reply != "" {
		t.Fatalf("Handle() = %q, %v", reply, err)
	}
}
