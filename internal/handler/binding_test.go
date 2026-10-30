package handler

import (
	"context"
	"strings"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type fakeBindingAPI struct {
	startCalls  []newapi.StartBindingRequest
	verifyCalls []newapi.VerifyBindingRequest
	startResult newapi.StartBindingResponse
	verifyErr   error
}

func (f *fakeBindingAPI) StartBinding(_ context.Context, request newapi.StartBindingRequest) (newapi.StartBindingResponse, error) {
	f.startCalls = append(f.startCalls, request)
	return f.startResult, nil
}

func (f *fakeBindingAPI) VerifyBinding(_ context.Context, request newapi.VerifyBindingRequest) (newapi.VerifyBindingResponse, error) {
	f.verifyCalls = append(f.verifyCalls, request)
	return newapi.VerifyBindingResponse{Username: "demo-user"}, f.verifyErr
}

func TestBindingHandlerStartsEmailVerificationAndCompletesIt(t *testing.T) {
	api := &fakeBindingAPI{startResult: newapi.StartBindingResponse{FlowToken: "flow-token"}}
	h := NewBindingHandler(api)

	prompt, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "绑定账号"})
	if err != nil || prompt == "" {
		t.Fatalf("start binding: reply=%q err=%v", prompt, err)
	}
	reply, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "user@example.com"})
	if err != nil || reply != "验证码已发送到邮箱，请把邮件里的 6 位数字验证码发给我。" {
		t.Fatalf("email step: reply=%q err=%v", reply, err)
	}
	reply, err = h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "123456"})
	if err != nil || reply != "绑定成功，之后可以使用 QQ Bot 了。" {
		t.Fatalf("code step: reply=%q err=%v", reply, err)
	}
	if len(api.verifyCalls) != 1 || api.verifyCalls[0].FlowToken != "flow-token" {
		t.Fatalf("verify calls = %+v", api.verifyCalls)
	}
}

func TestBindingHandlerConfirmsReplacementForExistingNonQQEmail(t *testing.T) {
	api := &fakeBindingAPI{startResult: newapi.StartBindingResponse{RequiresReplace: true, ExistingDomain: "example.com"}}
	h := NewBindingHandler(api)
	_, _ = h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "绑定账号"})
	reply, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "new@qq.com"})
	if err != nil || reply != "检测到账号已绑定非 QQ 邮箱（example.com），如果不换绑，后续无法使用 Bot。是否解绑后重新绑定？请回复：是/否" {
		t.Fatalf("replace prompt: reply=%q err=%v", reply, err)
	}
	api.startResult = newapi.StartBindingResponse{FlowToken: "replacement-flow"}
	reply, err = h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "是"})
	if err != nil || reply != "验证码已发送到邮箱，请把邮件里的 6 位数字验证码发给我。" {
		t.Fatalf("replace confirm: reply=%q err=%v", reply, err)
	}
	if len(api.startCalls) != 2 || !api.startCalls[1].ForceReplace {
		t.Fatalf("start calls = %+v", api.startCalls)
	}
}

func TestBindingHandlerKeepsGroupMemberSessionsIndependent(t *testing.T) {
	api := &fakeBindingAPI{startResult: newapi.StartBindingResponse{FlowToken: "flow-token"}}
	h := NewBindingHandler(api)
	for _, member := range []string{"member-1", "member-2"} {
		_, _ = h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: member, Content: "绑定账号"})
		reply, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: member, Content: member + "@example.com"})
		if err != nil || reply == "" {
			t.Fatalf("member %s email step: reply=%q err=%v", member, reply, err)
		}
	}
	if len(api.startCalls) != 2 {
		t.Fatalf("start calls = %+v", api.startCalls)
	}
	if api.startCalls[0].MemberOpenID == api.startCalls[1].MemberOpenID {
		t.Fatal("binding sessions shared the same member identity")
	}
}

func TestBindingHandlerRestartsWhenBindingCommandIsSentAgain(t *testing.T) {
	api := &fakeBindingAPI{startResult: newapi.StartBindingResponse{FlowToken: "flow-token"}}
	h := NewBindingHandler(api)
	start, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "绑定账号"})
	if err != nil || !strings.Contains(start, "邮箱") {
		t.Fatalf("initial Handle() = %q, %v", start, err)
	}
	if _, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "old@example.com"}); err != nil {
		t.Fatal(err)
	}
	restart, err := h.Handle(context.Background(), qq.GroupMessage{GroupOpenID: "group-1", MemberOpenID: "member-1", Content: "绑定账号"})
	if err != nil || !strings.Contains(restart, "邮箱") {
		t.Fatalf("restart Handle() = %q, %v", restart, err)
	}
}
