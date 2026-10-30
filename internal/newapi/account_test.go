package newapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAccountMethodsUseBotEndpoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/bot/v1/account/lookup":
			_, _ = w.Write([]byte(`{"success":true,"data":{"username":"demo-user","quota":12345,"quota_display":"¥0.01"}}`))
		case "/api/bot/v1/checkin":
			_, _ = w.Write([]byte(`{"success":true,"data":{"username":"demo-user","quota_awarded":100,"checkin_date":"2026-10-29"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient(ClientConfig{BaseURL: server.URL, BotID: "bot", Secret: "secret"})
	account, err := client.LookupAccount(context.Background(), AccountRequest{MemberOpenID: "member-1"})
	if err != nil || account.Username != "demo-user" || account.Quota != 12345 {
		t.Fatalf("LookupAccount() = %+v, %v", account, err)
	}
	checkin, err := client.CheckIn(context.Background(), AccountRequest{MemberOpenID: "member-1"})
	if err != nil || checkin.QuotaAwarded != 100 {
		t.Fatalf("CheckIn() = %+v, %v", checkin, err)
	}
}

func TestClientCheckInReportsDisabledFeature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false,"message":"Check-in feature is not enabled"}`))
	}))
	defer server.Close()

	client := NewClient(ClientConfig{BaseURL: server.URL, BotID: "bot", Secret: "secret"})
	if _, err := client.CheckIn(context.Background(), AccountRequest{MemberOpenID: "member-1"}); !errors.Is(err, ErrCheckInDisabled) {
		t.Fatalf("CheckIn() error = %v, want ErrCheckInDisabled", err)
	}
}

func TestClientCheckInReportsAlreadyCheckedIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"message":"Already checked in today"}`))
	}))
	defer server.Close()
	client := NewClient(ClientConfig{BaseURL: server.URL, BotID: "bot", Secret: "secret"})
	if _, err := client.CheckIn(context.Background(), AccountRequest{MemberOpenID: "member"}); !errors.Is(err, ErrAlreadyCheckedIn) {
		t.Fatalf("unexpected error: %v", err)
	}
}
