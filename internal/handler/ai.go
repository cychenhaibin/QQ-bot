package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

const aiSessionTTL = 30 * time.Minute
const aiMaxHistoryMessages = 40
const aiChoicesPerPage = 8

type AIAPI interface {
	ListAIGroups(context.Context, newapi.AIRequest) ([]newapi.AIGroup, error)
	GetAIAccess(context.Context, newapi.AIRequest) (newapi.AIAccessResponse, error)
	Chat(context.Context, string, string, []newapi.ChatMessage) (string, error)
}

type aiSessionStage uint8

const (
	aiAwaitGroup aiSessionStage = iota + 1
	aiAwaitModel
	aiActive
)

type aiSession struct {
	Page      int
	Stage     aiSessionStage
	Groups    []newapi.AIGroup
	Group     string
	Model     string
	APIKey    string
	Messages  []newapi.ChatMessage
	ExpiresAt time.Time
}

type AIHandler struct {
	api      AIAPI
	mu       sync.Mutex
	sessions map[string]aiSession
	locks    map[string]*sync.Mutex
	now      func() time.Time
}

func NewAIHandler(api AIAPI) *AIHandler {
	return &AIHandler{api: api, sessions: make(map[string]aiSession), locks: make(map[string]*sync.Mutex), now: time.Now}
}

func (h *AIHandler) Handle(ctx context.Context, event qq.GroupMessage) (string, error) {
	if event.Content != "AI 对话" && event.Content != "AI对话" && strings.TrimSpace(event.Content) != "停" {
		if !h.hasSession(event) {
			return "", nil
		}
	}
	if h.api == nil {
		return "AI 对话暂时不可用，请稍后重试。", errors.New("AI API is not configured")
	}
	key := event.GroupOpenID + "\x00" + event.MemberOpenID
	content := strings.TrimSpace(event.Content)
	unlock := h.lockSession(key)
	defer unlock()
	if content == "停" {
		if h.deleteSession(key) {
			return "本轮 AI 对话已停止。", nil
		}
		return "", nil
	}
	if content == "AI 对话" || content == "AI对话" {
		groups, err := h.api.ListAIGroups(ctx, newapi.AIRequest{MemberOpenID: event.MemberOpenID})
		if err != nil {
			return "AI 对话暂时无法开始，请先绑定账号或稍后重试。", err
		}
		if len(groups) == 0 {
			return "当前没有可用的 AI 分组。", nil
		}
		h.put(key, aiSession{Stage: aiAwaitGroup, Groups: groups})
		return formatAIGroupPrompt(groups), nil
	}
	session, ok := h.get(key)
	if !ok {
		return "", nil
	}
	if (session.Stage == aiAwaitGroup || session.Stage == aiAwaitModel) && (content == "上一页" || content == "下一页") {
		count := len(session.Groups)
		if session.Stage == aiAwaitModel {
			count = len(findModels(session.Groups, session.Group))
		}
		if content == "上一页" {
			session.Page--
		} else {
			session.Page++
		}
		session.Page = max(0, min(session.Page, max(0, (count-1)/aiChoicesPerPage)))
		h.put(key, session)
		if session.Stage == aiAwaitGroup {
			return formatAIGroupPage(session.Groups, session.Page), nil
		}
		return formatAIModelPage(newapi.AIGroup{Group: session.Group, Models: findModels(session.Groups, session.Group)}, session.Page), nil
	}
	switch session.Stage {
	case aiAwaitGroup:
		for _, group := range session.Groups {
			if group.Group != content {
				continue
			}
			access, err := h.api.GetAIAccess(ctx, newapi.AIRequest{MemberOpenID: event.MemberOpenID, Group: group.Group})
			if err != nil {
				return "AI 分组初始化失败，请稍后重试。", err
			}
			session.Stage, session.Group, session.APIKey = aiAwaitModel, group.Group, access.APIKey
			session.Page = 0
			h.put(key, session)
			return formatAIModelPrompt(group), nil
		}
		return "请发送列表中的分组名称。", nil
	case aiAwaitModel:
		if !containsString(findModels(session.Groups, session.Group), content) {
			return "请发送列表中的模型名称。", nil
		}
		session.Stage, session.Model = aiActive, content
		return h.chat(ctx, key, session, "hi")
	case aiActive:
		return h.chat(ctx, key, session, content)
	default:
		h.deleteSession(key)
		return "本轮 AI 对话已失效，请重新发送“AI 对话”。", nil
	}
}

func (h *AIHandler) chat(ctx context.Context, key string, session aiSession, content string) (string, error) {
	if len(session.Messages) >= aiMaxHistoryMessages {
		session.Messages = session.Messages[2:]
	}
	session.Messages = append(session.Messages, newapi.ChatMessage{Role: "user", Content: content})
	answer, err := h.api.Chat(ctx, session.APIKey, session.Model, session.Messages)
	if err != nil {
		return "AI 对话请求失败，请稍后重试。", err
	}
	session.Messages = append(session.Messages, newapi.ChatMessage{Role: "assistant", Content: answer})
	h.put(key, session)
	return answer, nil
}

func (h *AIHandler) lockSession(key string) func() {
	h.mu.Lock()
	lock, ok := h.locks[key]
	if !ok {
		lock = &sync.Mutex{}
		h.locks[key] = lock
	}
	h.mu.Unlock()
	lock.Lock()
	return lock.Unlock
}

func (h *AIHandler) get(key string) (aiSession, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	session, ok := h.sessions[key]
	if ok && !h.now().Before(session.ExpiresAt) {
		delete(h.sessions, key)
		return aiSession{}, false
	}
	return session, ok
}

func (h *AIHandler) put(key string, session aiSession) {
	session.ExpiresAt = h.now().Add(aiSessionTTL)
	h.mu.Lock()
	h.sessions[key] = session
	h.mu.Unlock()
}

func (h *AIHandler) deleteSession(key string) bool {
	h.mu.Lock()
	_, ok := h.sessions[key]
	delete(h.sessions, key)
	h.mu.Unlock()
	return ok
}

func (h *AIHandler) hasSession(event qq.GroupMessage) bool {
	_, ok := h.get(event.GroupOpenID + "\x00" + event.MemberOpenID)
	return ok
}

func formatAIGroupPrompt(groups []newapi.AIGroup) string {
	return formatAIGroupPage(groups, 0)
}

func formatAIGroupPage(groups []newapi.AIGroup, page int) string {
	names := make([]string, 0, len(groups))
	for _, group := range groups {
		names = append(names, group.Group)
	}
	choices, footer := formatAIChoicePage(names, page)
	return "请选择使用哪个分组：" + choices + "\n未在 API 密钥中创建的分组将自动创建。发送“停”将停止本轮对话。" + footer
}

func formatAIModelPrompt(group newapi.AIGroup) string {
	return formatAIModelPage(group, 0)
}

func formatAIModelPage(group newapi.AIGroup, page int) string {
	choices, footer := formatAIChoicePage(group.Models, page)
	return group.Group + "分组下有以下模型请先选择：" + choices + footer
}

func formatAIChoicePage(choices []string, page int) (string, string) {
	pages := max(1, (len(choices)+aiChoicesPerPage-1)/aiChoicesPerPage)
	page = max(0, min(page, pages-1))
	start := min(page*aiChoicesPerPage, len(choices))
	end := min(start+aiChoicesPerPage, len(choices))
	footer := ""
	if pages > 1 {
		footer = fmt.Sprintf("\n第 %d/%d 页", page+1, pages)
		if page > 0 {
			footer += "\n[上一页]"
		}
		if page+1 < pages {
			footer += "\n[下一页]"
		}
	}
	return strings.Join(choices[start:end], "、"), footer
}

func findModels(groups []newapi.AIGroup, name string) []string {
	for _, group := range groups {
		if group.Group == name {
			return group.Models
		}
	}
	return nil
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
