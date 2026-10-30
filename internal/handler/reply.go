package handler

import (
	"context"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type ReplySender interface {
	SendGroupMessage(context.Context, string, string, string) error
	SendGroupKeyboardMessage(context.Context, string, string, qq.Keyboard, string) error
	SendGroupMessageForInteraction(context.Context, string, string, string) error
	SendGroupKeyboardMessageForInteraction(context.Context, string, string, qq.Keyboard, string) error
}

func SendGroupReply(ctx context.Context, sender ReplySender, groupOpenID, content, messageID, interactionID string) error {
	if keyboard, ok := KeyboardForPrompt(content); ok {
		var err error
		if interactionID != "" {
			err = sender.SendGroupKeyboardMessageForInteraction(ctx, groupOpenID, content, keyboard, interactionID)
		} else {
			err = sender.SendGroupKeyboardMessage(ctx, groupOpenID, content, keyboard, messageID)
		}
		if err == nil {
			return nil
		}
	}
	if interactionID != "" {
		return sender.SendGroupMessageForInteraction(ctx, groupOpenID, content, interactionID)
	}
	return sender.SendGroupMessage(ctx, groupOpenID, content, messageID)
}
