package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

type ChannelAPI interface {
	ListChannels(context.Context) ([]newapi.ChannelGroup, error)
}

type ChannelHandler struct {
	api ChannelAPI
}

func NewChannelHandler(api ChannelAPI) *ChannelHandler {
	return &ChannelHandler{api: api}
}

func (h *ChannelHandler) Handle(ctx context.Context, event qq.GroupMessage) (string, error) {
	if event.Content != "模型渠道" {
		return "", nil
	}
	if h.api == nil {
		return "模型渠道暂时不可用，请稍后重试。", errors.New("channel API is not configured")
	}
	groups, err := h.api.ListChannels(ctx)
	if err != nil {
		return "模型渠道暂时不可用，请稍后重试。", err
	}
	if len(groups) == 0 {
		return "当前没有可用的模型渠道。", nil
	}
	var builder strings.Builder
	builder.WriteString("当前可用的模型渠道：")
	for _, group := range groups {
		builder.WriteString("\n\n分组：")
		builder.WriteString(group.Group)
		builder.WriteString("\n模型：")
		builder.WriteString(strings.Join(group.Models, "、"))
	}
	return builder.String(), nil
}
