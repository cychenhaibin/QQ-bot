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
func (s *replySender) SendGroupKeyboardMessageForInteraction(context.Context, string, string, qq.Keyboard, string) error {
	s.keyboardCalls++
	return s.keyboardErr
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
