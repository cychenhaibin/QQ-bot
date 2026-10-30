package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type replySender struct {
	keyboardCalls int
	plainCalls    int
	keyboardErr   error
	content       string
	eventID       string
}

func (s *replySender) SendGroupMessage(context.Context, string, string, string) error {
	s.plainCalls++
	return nil
}
func (s *replySender) SendGroupKeyboardMessage(context.Context, string, string, qq.Keyboard, string) error {
	s.keyboardCalls++
	return s.keyboardErr
}
func (s *replySender) SendGroupMessageForInteraction(context.Context, string, string, string) error {
	s.plainCalls++
	return nil
}
func (s *replySender) SendGroupKeyboardMessageForInteraction(_ context.Context, _ string, content string, _ qq.Keyboard, eventID string) error {
	s.keyboardCalls++
	s.content = content
	s.eventID = eventID
	return s.keyboardErr
}

func TestSendGroupReplyMentionsUserAndRemovesDuplicateSiteChoices(t *testing.T) {
	sender := &replySender{}
	err := SendGroupReply(context.Background(), sender, "group", siteSelectionPrompt(), "", "gateway-event", "member")
	if err != nil {
		t.Fatal(err)
	}
	if sender.content != "<@member> 请选择要使用的站点：" || sender.eventID != "gateway-event" || sender.plainCalls != 0 {
		t.Fatalf("unexpected keyboard reply: %+v", sender)
	}
}

func TestSendGroupReplyFallsBackToPlainTextWhenKeyboardFails(t *testing.T) {
	sender := &replySender{keyboardErr: errors.New("keyboard unsupported")}
	if err := SendGroupReply(context.Background(), sender, "group", BuildMainMenu(), "message", ""); err != nil {
		t.Fatalf("SendGroupReply() error = %v", err)
	}
	if sender.keyboardCalls != 1 || sender.plainCalls != 1 {
		t.Fatalf("sender calls = keyboard:%d plain:%d", sender.keyboardCalls, sender.plainCalls)
	}
}
