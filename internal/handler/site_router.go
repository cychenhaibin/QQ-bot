package handler

import (
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
	"github.com/QuantumNous/qq-bot-service/internal/site"
)

const siteSelectionTTL = 30 * time.Minute

type siteOperation uint8

const (
	siteOperationOneShot siteOperation = iota + 1
	siteOperationBinding
	siteOperationAI
)

type siteSession struct {
	Site      site.ID
	Command   string
	Operation siteOperation
	ExpiresAt time.Time
}

type SiteDispatch func(site.ID, qq.GroupMessage) (string, error)

type SiteRouter struct {
	mu       sync.Mutex
	sessions map[string]siteSession
	now      func() time.Time
}

func NewSiteRouter() *SiteRouter {
	return &SiteRouter{sessions: make(map[string]siteSession), now: time.Now}
}

func (r *SiteRouter) Handle(event qq.GroupMessage, dispatch SiteDispatch) (string, bool, error) {
	key := event.GroupOpenID + "\x00" + event.MemberOpenID
	content := strings.TrimSpace(event.Content)
	session, exists := r.get(key)
	if exists && isOperationStart(content) {
		r.delete(key)
		exists = false
	}
	if exists && session.Site == "" {
		selected, ok := site.Parse(content)
		if !ok {
			return siteSelectionPrompt(), true, nil
		}
		session.Site = selected
		r.put(key, session)
		reply, err := dispatch(selected, eventWithContent(event, session.Command))
		if session.Operation == siteOperationOneShot {
			r.delete(key)
		}
		return reply, true, err
	}
	if exists {
		reply, err := dispatch(session.Site, event)
		return reply, true, err
	}
	operation, ok := operationFor(content)
	if !ok {
		return "", false, nil
	}
	r.put(key, siteSession{Command: content, Operation: operation})
	return siteSelectionPrompt(), true, nil
}

func (r *SiteRouter) get(key string) (siteSession, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[key]
	if ok && !r.now().Before(session.ExpiresAt) {
		delete(r.sessions, key)
		return siteSession{}, false
	}
	return session, ok
}

func (r *SiteRouter) put(key string, session siteSession) {
	session.ExpiresAt = r.now().Add(siteSelectionTTL)
	r.mu.Lock()
	r.sessions[key] = session
	r.mu.Unlock()
}

func (r *SiteRouter) delete(key string) {
	r.mu.Lock()
	delete(r.sessions, key)
	r.mu.Unlock()
}

func eventWithContent(event qq.GroupMessage, content string) qq.GroupMessage {
	event.Content = content
	return event
}

func siteSelectionPrompt() string {
	return "请选择要使用的站点：\n[FluxA] [Charity]"
}

func operationFor(content string) (siteOperation, bool) {
	switch content {
	case "绑定账号", "绑定":
		return siteOperationBinding, true
	case "AI 对话", "AI对话":
		return siteOperationAI, true
	case "查询余额", "每日签到", "充值中心", "模型渠道":
		return siteOperationOneShot, true
	default:
		return 0, false
	}
}

func isOperationStart(content string) bool {
	_, ok := operationFor(content)
	return ok
}
