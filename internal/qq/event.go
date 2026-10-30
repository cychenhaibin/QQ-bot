package qq

import (
	"encoding/json"
	"errors"
	"strings"
)

var errUnsupportedEvent = errors.New("unsupported QQ group message event")

type GroupMessage struct {
	MessageID    string
	GroupOpenID  string
	MemberOpenID string
	Content      string
}

type groupMessagePayload struct {
	ID          string `json:"id"`
	GroupOpenID string `json:"group_openid"`
	Content     string `json:"content"`
	Author      struct {
		MemberOpenID string `json:"member_openid"`
	} `json:"author"`
}

func DecodeGroupMessage(data []byte, eventType string) (GroupMessage, error) {
	if eventType != "GROUP_AT_MESSAGE_CREATE" && eventType != "GROUP_MESSAGE_CREATE" {
		return GroupMessage{}, errUnsupportedEvent
	}
	var payload groupMessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return GroupMessage{}, err
	}
	if strings.TrimSpace(payload.ID) == "" || strings.TrimSpace(payload.GroupOpenID) == "" || strings.TrimSpace(payload.Author.MemberOpenID) == "" {
		return GroupMessage{}, errors.New("QQ group message is missing identity fields")
	}
	return GroupMessage{
		MessageID:    payload.ID,
		GroupOpenID:  payload.GroupOpenID,
		MemberOpenID: payload.Author.MemberOpenID,
		Content:      strings.TrimSpace(payload.Content),
	}, nil
}
