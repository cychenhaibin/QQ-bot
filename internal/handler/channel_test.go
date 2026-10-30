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
	h := NewChannelHandler(fakeChannelAPI{groups: []newapi.ChannelGroup{{Group: "default", Models: []string{"gpt-4", "claude-3"}}}})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{Content: "模型渠道"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "default") || !strings.Contains(reply, "gpt-4") || !strings.Contains(reply, "claude-3") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestChannelHandlerIgnoresOtherMessages(t *testing.T) {
	h := NewChannelHandler(fakeChannelAPI{})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{Content: "充值中心"})
	if err != nil || reply != "" {
		t.Fatalf("Handle() = %q, %v", reply, err)
	}
}
