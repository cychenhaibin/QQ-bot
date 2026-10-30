package qq

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientGetsAccessTokenAndSendsGroupMessage(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/app/getAppAccessToken":
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"appId":"app-id"`) {
				t.Errorf("token request body = %s", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":7200}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/groups/group-1/messages":
			authHeader = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL,
		AppID:         "app-id",
		AppSecret:     "app-secret",
		TokenEndpoint: server.URL + "/app/getAppAccessToken",
	})

	if err := client.SendGroupMessage(context.Background(), "group-1", "hello", "message-1"); err != nil {
		t.Fatalf("SendGroupMessage() error = %v", err)
	}
	if authHeader != "QQBot access-token" {
		t.Fatalf("Authorization = %q", authHeader)
	}
}

func TestClientAcceptsStringExpiresInFromQQTokenEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":"7200"}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL,
		AppID:         "app-id",
		AppSecret:     "app-secret",
		TokenEndpoint: server.URL,
	})

	if token, err := client.AccessToken(context.Background()); err != nil || token != "access-token" {
		t.Fatalf("AccessToken() = %q, %v", token, err)
	}
}

func TestClientSendsGroupKeyboardMessage(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/app/getAppAccessToken":
			_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":7200}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/groups/group-1/messages":
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode message payload: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL,
		AppID:         "app-id",
		AppSecret:     "app-secret",
		TokenEndpoint: server.URL + "/app/getAppAccessToken",
	})
	keyboard := Keyboard{Content: KeyboardContent{Rows: []KeyboardRow{{Buttons: []KeyboardButton{{
		ID:         "menu_checkin",
		RenderData: RenderData{Label: "每日签到", VisitedLabel: "每日签到", Style: 1},
		Action:     ButtonAction{Type: 1, Permission: ButtonPermission{Type: 2}, Data: "每日签到", UnsupportTips: "暂不支持按钮操作"},
	}}}}}}

	if err := client.SendGroupKeyboardMessage(context.Background(), "group-1", "请选择功能：", keyboard, "message-1"); err != nil {
		t.Fatalf("SendGroupKeyboardMessage() error = %v", err)
	}
	if payload["msg_type"] != float64(2) || payload["msg_id"] != "message-1" {
		t.Fatalf("payload metadata = %#v", payload)
	}
	if payload["markdown"] == nil || payload["keyboard"] == nil {
		t.Fatalf("payload missing markdown or keyboard: %#v", payload)
	}
	if _, exists := payload["content"]; exists {
		t.Fatal("markdown keyboard message must not duplicate ordinary text content")
	}
	payload = nil
	if err := client.SendGroupKeyboardMessageForInteraction(context.Background(), "group-1", "请选择站点：", keyboard, "event-1"); err != nil {
		t.Fatal(err)
	}
	if _, exists := payload["content"]; exists || payload["event_id"] != "event-1" {
		t.Fatalf("unexpected interaction keyboard payload: %#v", payload)
	}
}

func TestClientAcknowledgesInteraction(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app/getAppAccessToken" {
			if r.Method != http.MethodPut || r.URL.Path != "/interactions/interaction-1" {
				http.NotFound(w, r)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode acknowledgement: %v", err)
			}
			if got := r.Header.Get("Authorization"); got != "QQBot access-token" {
				t.Errorf("Authorization = %q", got)
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":7200}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL,
		AppID:         "app-id",
		AppSecret:     "app-secret",
		TokenEndpoint: server.URL + "/app/getAppAccessToken",
	})
	if err := client.AcknowledgeInteraction(context.Background(), "interaction-1"); err != nil {
		t.Fatalf("AcknowledgeInteraction() error = %v", err)
	}
	if payload["code"] != float64(0) {
		t.Fatalf("acknowledgement payload = %#v", payload)
	}
}

func TestClientSendsGroupMessageForInteractionEvent(t *testing.T) {
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/app/getAppAccessToken":
			_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":7200}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v2/groups/group-1/messages":
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode message payload: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(ClientConfig{HTTPClient: server.Client(), APIBaseURL: server.URL, AppID: "app-id", AppSecret: "app-secret", TokenEndpoint: server.URL + "/app/getAppAccessToken"})
	if err := client.SendGroupMessageForInteraction(context.Background(), "group-1", "done", "interaction-1"); err != nil {
		t.Fatalf("SendGroupMessageForInteraction() error = %v", err)
	}
	if payload["event_id"] != "interaction-1" {
		t.Fatalf("payload event_id = %#v", payload["event_id"])
	}
}
