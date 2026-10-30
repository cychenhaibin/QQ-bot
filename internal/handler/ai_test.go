package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type fakeAIAPI struct{}

func (fakeAIAPI) ListAIGroups(context.Context, newapi.AIRequest) ([]newapi.AIGroup, error) {
	return []newapi.AIGroup{{Group: "default", Models: []string{"gpt-4"}}}, nil
}
func (fakeAIAPI) GetAIAccess(context.Context, newapi.AIRequest) (newapi.AIAccessResponse, error) {
	return newapi.AIAccessResponse{APIKey: "sk-test"}, nil
}
func (fakeAIAPI) Chat(context.Context, string, string, []newapi.ChatMessage) (string, error) {
	return "answer", nil
}

func TestAIHandlerKeepsIndependentUserSessionAndStopsOnCommand(t *testing.T) {
	h := NewAIHandler(fakeAIAPI{})
	event := qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member-1", Content: "AI 对话"}
	reply, err := h.Handle(context.Background(), event)
	if err != nil || !strings.Contains(reply, "default") {
		t.Fatalf("start = %q, %v", reply, err)
	}
	event.Content = "default"
	reply, err = h.Handle(context.Background(), event)
	if err != nil || !strings.Contains(reply, "gpt-4") {
		t.Fatalf("group = %q, %v", reply, err)
	}
	event.Content = "gpt-4"
	reply, err = h.Handle(context.Background(), event)
	if err != nil || reply != "answer" {
		t.Fatalf("model = %q, %v", reply, err)
	}
	event.Content = "停"
	reply, err = h.Handle(context.Background(), event)
	if err != nil || reply == "" {
		t.Fatalf("stop = %q, %v", reply, err)
	}
}
