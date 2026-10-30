package newapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientListsModelChannels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bot/v1/channels" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"groups":[{"group":"default","models":["gpt-4"]}]}}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{BaseURL: server.URL, BotID: "bot", Secret: "secret"})
	groups, err := client.ListChannels(context.Background())
	if err != nil || len(groups) != 1 || groups[0].Group != "default" {
		t.Fatalf("ListChannels() = %+v, %v", groups, err)
	}
}
