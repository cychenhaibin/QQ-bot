package newapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAIEndpointsAndChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/bot/v1/ai/groups":
			_, _ = w.Write([]byte(`{"success":true,"data":{"groups":[{"group":"default","models":["gpt-4"]}]}}`))
		case "/api/bot/v1/ai/access":
			_, _ = w.Write([]byte(`{"success":true,"data":{"api_key":"sk-test","group":"default"}}`))
		case "/v1/chat/completions":
			if r.Header.Get("Authorization") != "Bearer sk-test" {
				t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
			}
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(ClientConfig{BaseURL: server.URL, BotID: "bot", Secret: "secret"})
	groups, err := client.ListAIGroups(context.Background(), AIRequest{MemberOpenID: "member-1"})
	if err != nil || len(groups) != 1 {
		t.Fatalf("ListAIGroups() = %+v, %v", groups, err)
	}
	access, err := client.GetAIAccess(context.Background(), AIRequest{MemberOpenID: "member-1", Group: "default"})
	if err != nil || access.APIKey != "sk-test" {
		t.Fatalf("GetAIAccess() = %+v, %v", access, err)
	}
	content, err := client.Chat(context.Background(), access.APIKey, "gpt-4", []ChatMessage{{Role: "user", Content: "hi"}})
	if err != nil || content != "hello" {
		t.Fatalf("Chat() = %q, %v", content, err)
	}
}
