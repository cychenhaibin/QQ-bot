package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

func TestBuildMainMenuListsSupportedFeatures(t *testing.T) {
	menu := BuildMainMenu()
	for _, item := range []string{"每日签到", "查询余额", "绑定账号", "AI 对话", "充值中心", "模型渠道"} {
		if !strings.Contains(menu, item) {
			t.Fatalf("menu does not contain %q: %s", item, menu)
		}
	}
}

type menuSender struct {
	plainCalls   int
	keyboardCall *qq.Keyboard
}

func (s *menuSender) SendGroupMessage(context.Context, string, string, string) error {
	s.plainCalls++
	return nil
}

func (s *menuSender) SendGroupKeyboardMessage(_ context.Context, _ string, _ string, keyboard qq.Keyboard, _ string) error {
	s.keyboardCall = &keyboard
	return nil
}

func TestMenuHandlerSendsKeyboardButtons(t *testing.T) {
	sender := &menuSender{}
	h := MenuHandler{Sender: sender}
	if err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group", Content: "菜单"}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if sender.keyboardCall == nil || len(sender.keyboardCall.Content.Rows) != 3 || sender.plainCalls != 0 {
		t.Fatalf("sender calls = keyboard:%#v plain:%d", sender.keyboardCall, sender.plainCalls)
	}
}
