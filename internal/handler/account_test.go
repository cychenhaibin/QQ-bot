package handler

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type fakeAccountAPI struct {
	account newapi.AccountResponse
	checkin newapi.CheckinResponse
	err     error
}

func (f fakeAccountAPI) LookupAccount(context.Context, newapi.AccountRequest) (newapi.AccountResponse, error) {
	return f.account, f.err
}

func (f fakeAccountAPI) CheckIn(context.Context, newapi.AccountRequest) (newapi.CheckinResponse, error) {
	return f.checkin, f.err
}

func TestAccountHandlerRepliesWithUsernameAndBalance(t *testing.T) {
	h := NewAccountHandler(fakeAccountAPI{account: newapi.AccountResponse{Username: "demo", QuotaDisplay: "¥1.23"}})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "查询余额"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reply, "用户名：demo") || !strings.Contains(reply, "¥1.23") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestAccountHandlerRepliesWithExactCheckinSuccessMessage(t *testing.T) {
	h := NewAccountHandler(fakeAccountAPI{checkin: newapi.CheckinResponse{QuotaAwarded: 100}})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "每日签到"})
	if err != nil || reply != "✅ 签到成功，今日奖励已到账。" {
		t.Fatalf("Handle() = %q, %v", reply, err)
	}
}

func TestAccountHandlerExplainsWhenCheckinIsDisabled(t *testing.T) {
	h := NewAccountHandler(fakeAccountAPI{err: newapi.ErrCheckInDisabled})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "每日签到"})
	if err != nil || reply != "当前站点未开启签到功能，暂时无法签到。" {
		t.Fatalf("Handle() = %q, %v", reply, err)
	}
}

func TestAccountHandlerDoesNotReplyForOtherMessages(t *testing.T) {
	h := NewAccountHandler(fakeAccountAPI{err: errors.New("must not call")})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{Content: "菜单"})
	if err != nil || reply != "" {
		t.Fatalf("Handle() = %q, %v", reply, err)
	}
}
