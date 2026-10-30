package handler

import (
	"context"
	"strings"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

func BuildMainMenu() string {
	return "请选择功能：\n" +
		"[每日签到] [查询余额]\n" +
		"[绑定账号] [AI 对话]\n" +
		"[充值中心] [模型渠道]"
}

type GroupMessageSender interface {
	SendGroupMessage(context.Context, string, string, string) error
}

type GroupKeyboardSender interface {
	GroupMessageSender
	SendGroupKeyboardMessage(context.Context, string, string, qq.Keyboard, string) error
}

type MenuHandler struct {
	Sender GroupMessageSender
}

func (h MenuHandler) Handle(ctx context.Context, event qq.GroupMessage) error {
	if !IsMenuCommand(event.Content) {
		return nil
	}
	menu := BuildMainMenu()
	if sender, ok := h.Sender.(GroupKeyboardSender); ok {
		if keyboard, hasKeyboard := KeyboardForPrompt(menu); hasKeyboard {
			prompt := KeyboardPrompt(menu)
			if event.MemberOpenID != "" {
				prompt = "<@" + event.MemberOpenID + "> " + prompt
			}
			if err := sender.SendGroupKeyboardMessage(ctx, event.GroupOpenID, prompt, keyboard, event.MessageID); err == nil {
				return nil
			}
		}
	}
	return h.Sender.SendGroupMessage(ctx, event.GroupOpenID, menu, event.MessageID)
}

func IsMenuCommand(content string) bool {
	switch strings.TrimSpace(content) {
	case "菜单", "/菜单", "帮助":
		return true
	default:
		return false
	}
}
