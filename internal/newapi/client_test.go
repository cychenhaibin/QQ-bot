package newapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientSignsInternalRequestWithoutExposingSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		timestamp := r.Header.Get("X-Bot-Timestamp")
		nonce := r.Header.Get("X-Bot-Nonce")
		mac := hmac.New(sha256.New, []byte("service-secret"))
		_, _ = mac.Write([]byte("POST\n/api/bot/v1/checkin\n" + timestamp + "\n" + nonce + "\n" + string(body)))
		if got, want := r.Header.Get("X-Bot-Signature"), hex.EncodeToString(mac.Sum(nil)); got != want {
			t.Fatalf("signature = %q, want %q", got, want)
		}
		if got := r.Header.Get("Authorization"); strings.Contains(got, "service-secret") {
			t.Fatal("service secret was sent in Authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{
		HTTPClient: server.Client(),
		BaseURL:    server.URL,
		BotID:      "qq-bot",
		Secret:     "service-secret",
		Now:        func() time.Time { return time.Unix(1700000000, 0) },
		Nonce:      func() (string, error) { return "nonce-1", nil },
	})
	var response struct {
		Success bool `json:"success"`
	}
	if err := client.DoJSON(context.Background(), http.MethodPost, "/api/bot/v1/checkin", map[string]string{"member_openid": "member-1"}, &response); err != nil {
		t.Fatalf("DoJSON() error = %v", err)
	}
	if !response.Success {
		t.Fatal("response.Success = false")
	}
}
