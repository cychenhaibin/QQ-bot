package newapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientBindingMethodsUseBotEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/bot/v1/bindings/start" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"flow_token":"flow-1"}}`))
			return
		}
		if r.URL.Path == "/api/bot/v1/bindings/verify" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"data":{"username":"demo-user"}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewClient(ClientConfig{BaseURL: server.URL, BotID: "bot", Secret: "secret"})
	start, err := client.StartBinding(context.Background(), StartBindingRequest{MemberOpenID: "member-1", Email: "new@qq.com"})
	if err != nil || start.FlowToken != "flow-1" {
		t.Fatalf("StartBinding() = %+v, %v", start, err)
	}
	verified, err := client.VerifyBinding(context.Background(), VerifyBindingRequest{MemberOpenID: "member-1", FlowToken: "flow-1", Code: "123456"})
	if err != nil || verified.Username != "demo-user" {
		t.Fatalf("VerifyBinding() = %+v, %v", verified, err)
	}
}
