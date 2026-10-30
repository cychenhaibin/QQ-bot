package handler

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

const bindingSessionTTL = 10 * time.Minute

type BindingAPI interface {
	StartBinding(context.Context, newapi.StartBindingRequest) (newapi.StartBindingResponse, error)
	VerifyBinding(context.Context, newapi.VerifyBindingRequest) (newapi.VerifyBindingResponse, error)
}

type bindingStage uint8

const (
	bindingStageAwaitEmail bindingStage = iota + 1
	bindingStageAwaitReplace
	bindingStageAwaitCode
)

type bindingSession struct {
	Stage     bindingStage
	Email     string
	FlowToken string
	ExpiresAt time.Time
}

type BindingHandler struct {
	api      BindingAPI
	mu       sync.Mutex
	sessions map[string]bindingSession
	now      func() time.Time
}

func NewBindingHandler(api BindingAPI) *BindingHandler {
	return &BindingHandler{api: api, sessions: make(map[string]bindingSession), now: time.Now}
}

func (h *BindingHandler) Handle(ctx context.Context, event qq.GroupMessage) (string, error) {
	if h.api == nil {
		return "", errors.New("binding API is not configured")
	}
	key := event.GroupOpenID + "\x00" + event.MemberOpenID
	content := strings.TrimSpace(event.Content)
	if content == "绑定账号" || content == "绑定" {
		h.delete(key)
		h.put(key, bindingSession{Stage: bindingStageAwaitEmail})
		return "请输入要绑定的邮箱地址，例如：xxxxx@qq.com", nil
	}
	h.mu.Lock()
	session, exists := h.sessions[key]
	if exists && !h.now().Before(session.ExpiresAt) {
		delete(h.sessions, key)
		exists = false
	}
	h.mu.Unlock()

	if !exists {
		return "", nil
	}

	switch session.Stage {
	case bindingStageAwaitEmail:
		return h.handleEmail(ctx, key, event.MemberOpenID, content)
	case bindingStageAwaitReplace:
		return h.handleReplace(ctx, key, event.MemberOpenID, session.Email, content)
	case bindingStageAwaitCode:
		return h.handleCode(ctx, key, event.MemberOpenID, session.FlowToken, content)
	default:
		h.delete(key)
		return "绑定流程已失效，请重新发送“绑定账号”。", nil
	}
}

func (h *BindingHandler) handleEmail(ctx context.Context, key, memberOpenID, email string) (string, error) {
	if !validEmail(email) {
		return "邮箱格式不正确，请重新输入。", nil
	}
	result, err := h.api.StartBinding(ctx, newapi.StartBindingRequest{MemberOpenID: memberOpenID, Email: email})
	if err != nil {
		return "绑定暂时无法开始，请稍后重试。", err
	}
	if result.RequiresReplace {
		domain := result.ExistingDomain
		if domain == "" {
			domain = "其他域名"
		}
		h.put(key, bindingSession{Stage: bindingStageAwaitReplace, Email: email})
		return "检测到账号已绑定非 QQ 邮箱（" + domain + "），如果不换绑，后续无法使用 Bot。是否解绑后重新绑定？请回复：是/否", nil
	}
	if result.FlowToken == "" {
		return "绑定暂时无法开始，请稍后重试。", errors.New("new-api returned an empty binding flow token")
	}
	h.put(key, bindingSession{Stage: bindingStageAwaitCode, Email: email, FlowToken: result.FlowToken})
	return "验证码已发送到邮箱，请把邮件里的 6 位数字验证码发给我。", nil
}

func (h *BindingHandler) handleReplace(ctx context.Context, key, memberOpenID, email, answer string) (string, error) {
	switch answer {
	case "否", "不", "no", "N", "n":
		h.delete(key)
		return "已取消换绑。", nil
	case "是", "好", "yes", "Y", "y":
		result, err := h.api.StartBinding(ctx, newapi.StartBindingRequest{MemberOpenID: memberOpenID, Email: email, ForceReplace: true})
		if err != nil {
			return "换绑暂时无法开始，请稍后重试。", err
		}
		if result.FlowToken == "" {
			return "换绑暂时无法开始，请稍后重试。", errors.New("new-api returned an empty replacement flow token")
		}
		h.put(key, bindingSession{Stage: bindingStageAwaitCode, Email: email, FlowToken: result.FlowToken})
		return "验证码已发送到邮箱，请把邮件里的 6 位数字验证码发给我。", nil
	default:
		return "请回复“是”或“否”。", nil
	}
}

func (h *BindingHandler) handleCode(ctx context.Context, key, memberOpenID, flowToken, code string) (string, error) {
	if len(code) != 6 || strings.IndexFunc(code, func(r rune) bool { return !unicode.IsDigit(r) }) >= 0 {
		return "验证码格式不正确，请发送 6 位数字验证码。", nil
	}
	if _, err := h.api.VerifyBinding(ctx, newapi.VerifyBindingRequest{MemberOpenID: memberOpenID, FlowToken: flowToken, Code: code}); err != nil {
		return "验证码错误或已过期，请重新发送验证码。", err
	}
	h.delete(key)
	return "绑定成功，之后可以使用 QQ Bot 了。", nil
}

func (h *BindingHandler) put(key string, session bindingSession) {
	session.ExpiresAt = h.now().Add(bindingSessionTTL)
	h.mu.Lock()
	h.sessions[key] = session
	h.mu.Unlock()
}

func (h *BindingHandler) delete(key string) {
	h.mu.Lock()
	delete(h.sessions, key)
	h.mu.Unlock()
}

func validEmail(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address == value && strings.Contains(value, "@")
}
