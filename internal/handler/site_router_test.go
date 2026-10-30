package handler

import (
	"errors"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
	"github.com/QuantumNous/qq-bot-service/internal/site"
)

func TestSiteRouterAsksForSiteBeforeDispatchingAndRoutesSelection(t *testing.T) {
	router := NewSiteRouter()
	var calls []struct {
		site    site.ID
		content string
	}
	dispatch := func(selected site.ID, event qq.GroupMessage) (string, error) {
		calls = append(calls, struct {
			site    site.ID
			content string
		}{selected, event.Content})
		return "done", nil
	}

	reply, handled, err := router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "查询余额"}, dispatch)
	if err != nil || !handled || reply != "请选择要使用的站点：\n[FluxA] [Charity]" {
		t.Fatalf("initial Handle() = %q, %v, %v", reply, handled, err)
	}

	reply, handled, err = router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "FluxA"}, dispatch)
	if err != nil || !handled || reply != "done" {
		t.Fatalf("selection Handle() = %q, %v, %v", reply, handled, err)
	}
	if len(calls) != 1 || calls[0].site != site.FluxA || calls[0].content != "查询余额" {
		t.Fatalf("dispatch calls = %+v", calls)
	}
}

func TestSiteRouterKeepsSelectedSiteForMultiStepFlow(t *testing.T) {
	router := NewSiteRouter()
	var calls []struct {
		site    site.ID
		content string
	}
	dispatch := func(selected site.ID, event qq.GroupMessage) (string, error) {
		calls = append(calls, struct {
			site    site.ID
			content string
		}{selected, event.Content})
		return "ok", nil
	}

	_, _, _ = router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "绑定账号"}, dispatch)
	_, _, _ = router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "Charity"}, dispatch)
	_, handled, err := router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "user@example.com"}, dispatch)
	if err != nil || !handled {
		t.Fatalf("follow-up Handle() = %v, %v", handled, err)
	}
	if len(calls) != 2 || calls[0].site != site.Charity || calls[1].site != site.Charity || calls[1].content != "user@example.com" {
		t.Fatalf("dispatch calls = %+v", calls)
	}
}

func TestSiteRouterAsksAgainForEachNewOperation(t *testing.T) {
	router := NewSiteRouter()
	dispatch := func(site.ID, qq.GroupMessage) (string, error) {
		return "done", nil
	}

	_, _, _ = router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "查询余额"}, dispatch)
	_, _, _ = router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "FluxA"}, dispatch)
	reply, handled, err := router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "每日签到"}, dispatch)
	if err != nil || !handled || reply != "请选择要使用的站点：\n[FluxA] [Charity]" {
		t.Fatalf("second operation Handle() = %q, %v, %v", reply, handled, err)
	}
}

func TestSiteRouterPropagatesDispatchError(t *testing.T) {
	router := NewSiteRouter()
	wantErr := errors.New("dispatch failed")
	dispatch := func(site.ID, qq.GroupMessage) (string, error) {
		return "failed", wantErr
	}
	_, _, _ = router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "查询余额"}, dispatch)
	reply, handled, err := router.Handle(qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "Charity"}, dispatch)
	if !handled || reply != "failed" || !errors.Is(err, wantErr) {
		t.Fatalf("error Handle() = %q, %v, %v", reply, handled, err)
	}
}
