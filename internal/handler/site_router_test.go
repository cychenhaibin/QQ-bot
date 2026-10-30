package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
	"github.com/QuantumNous/qq-bot-service/internal/site"
)

func TestSiteRouterStopClearsAIAndSiteSessionsAndAllowsMenu(t *testing.T) {
	router := NewSiteRouter()
	ai := NewAIHandler(fakeAIAPI{})
	event := qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member"}
	dispatch := func(_ site.ID, event qq.GroupMessage) (string, error) { return ai.Handle(context.Background(), event) }
	for _, command := range []string{"AI 对话", "FluxA", "default", "gpt-4"} {
		event.Content = command
		_, _, err := router.Handle(event, dispatch)
		if err != nil {
			t.Fatal(err)
		}
	}
	event.Content = "停"
	reply, handled, err := router.Handle(event, dispatch)
	if err != nil || !handled || reply != "本轮 AI 对话已停止。" {
		t.Fatalf("stop=%q %v %v", reply, handled, err)
	}
	if ai.hasSession(event) {
		t.Fatal("AI session survived stop")
	}
	for _, content := range []string{"菜单", "/菜单", "帮助", "hi"} {
		event.Content = content
		_, handled, err := router.Handle(event, dispatch)
		if err != nil || handled {
			t.Fatalf("post-stop %q swallowed: handled=%v err=%v", content, handled, err)
		}
	}
	event.Content = "AI 对话"
	reply, _, _ = router.Handle(event, dispatch)
	if reply != siteSelectionPrompt() {
		t.Fatalf("restart should ask for site: %q", reply)
	}
}

func TestSiteRouterMenuBypassesPendingAndActiveFlows(t *testing.T) {
	for _, flow := range []string{"AI 对话", "绑定账号"} {
		for _, selectSite := range []bool{false, true} {
			router := NewSiteRouter()
			event := qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: flow}
			dispatch := func(site.ID, qq.GroupMessage) (string, error) { return "started", nil }
			_, _, _ = router.Handle(event, dispatch)
			if selectSite {
				event.Content = "FluxA"
				_, _, _ = router.Handle(event, dispatch)
			}
			for _, menu := range []string{"菜单", "/菜单", "帮助"} {
				event.Content = menu
				_, handled, err := router.Handle(event, func(site.ID, qq.GroupMessage) (string, error) {
					t.Fatal("menu dispatched into active flow")
					return "", nil
				})
				if handled || err != nil {
					t.Fatalf("menu swallowed in %s: handled=%v err=%v", flow, handled, err)
				}
			}
		}
	}
}

func TestSiteRouterStopCancelsAIWhileAwaitingSite(t *testing.T) {
	router := NewSiteRouter()
	event := qq.GroupMessage{GroupOpenID: "group", MemberOpenID: "member", Content: "AI 对话"}
	dispatch := func(site.ID, qq.GroupMessage) (string, error) {
		t.Fatal("must not dispatch without site")
		return "", nil
	}
	_, _, _ = router.Handle(event, dispatch)
	event.Content = "停"
	reply, handled, err := router.Handle(event, dispatch)
	if err != nil || !handled || reply != "本轮 AI 对话已停止。" {
		t.Fatalf("stop=%q %v %v", reply, handled, err)
	}
	event.Content = "FluxA"
	_, handled, _ = router.Handle(event, dispatch)
	if handled {
		t.Fatal("pending site selection survived stop")
	}
}

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
