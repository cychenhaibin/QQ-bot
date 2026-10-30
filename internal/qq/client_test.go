package qq

import (
	"context"
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
