package handler

import (
	"context"
	"fmt"
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

type paginatedAIAPI struct{ fakeAIAPI }

func (paginatedAIAPI) ListAIGroups(context.Context, newapi.AIRequest) ([]newapi.AIGroup, error) {
	models := make([]string, 0, 17)
	for i := range 17 {
		models = append(models, fmt.Sprintf("model-%02d", i+1))
	}
	groups := make([]newapi.AIGroup, 0, 17)
	for i := range 17 {
		groups = append(groups, newapi.AIGroup{Group: fmt.Sprintf("group-%02d", i+1), Models: models})
	}
	return groups, nil
}

func TestAIChoicePagesUseEightChoicesWithTwoColumnsAndOneNavigationRow(t *testing.T) {
	h := NewAIHandler(paginatedAIAPI{})
	event := qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member"}
	commands := []struct{ command, want string }{
		{"AI 对话", "group-01"},
		{"上一页", "group-01"},
		{"下一页", "group-09"},
		{"下一页", "group-17"},
		{"下一页", "group-17"},
		{"上一页", "group-09"},
		{"group-09", "model-01"},
		{"下一页", "model-09"},
		{"下一页", "model-17"},
	}
	for _, step := range commands {
		event.Content = step.command
		prompt, err := h.Handle(context.Background(), event)
		if err != nil {
			t.Fatal(err)
		}
		keyboard, ok := KeyboardForPrompt(prompt)
		if !ok || len(keyboard.Content.Rows) > 5 || len(keyboard.Content.Rows) == 0 {
			t.Fatalf("invalid keyboard for %q: %+v", step.command, keyboard)
		}
		choiceCount := 0
		for index, row := range keyboard.Content.Rows {
			if index == len(keyboard.Content.Rows)-1 {
				if len(row.Buttons) > 3 || row.Buttons[len(row.Buttons)-1].Action.Data != "停" {
					t.Fatalf("invalid navigation row: %+v", row)
				}
				for _, button := range row.Buttons {
					if button.Action.Data != "上一页" && button.Action.Data != "下一页" && button.Action.Data != "停" {
						t.Fatalf("choice leaked into navigation: %+v", row)
					}
				}
				continue
			}
			if len(row.Buttons) > 2 || len(row.Buttons) == 0 {
				t.Fatalf("expected at most two buttons per row: %+v", row)
			}
			for _, button := range row.Buttons {
				if button.Action.Data != "上一页" && button.Action.Data != "下一页" && button.Action.Data != "停" {
					choiceCount++
				}
			}
		}
		wantCount := 8
		if strings.HasSuffix(step.want, "17") {
			wantCount = 1
		}
		if choiceCount != wantCount {
			t.Fatalf("choice count = %d, want %d", choiceCount, wantCount)
		}
		if got := keyboard.Content.Rows[0].Buttons[0].Action.Data; got != step.want {
			t.Fatalf("after %q first choice = %q, want %q", step.command, got, step.want)
		}
	}
	event.Content = "model-17"
	answer, err := h.Handle(context.Background(), event)
	if err != nil || answer != "answer" {
		t.Fatalf("model selection: %q %v", answer, err)
	}
}
