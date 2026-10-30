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

type Interaction struct {
	ID           string
	GroupOpenID  string
	MemberOpenID string
	ButtonData   string
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
	content := strings.TrimSpace(payload.Content)
	if eventType == "GROUP_AT_MESSAGE_CREATE" && content == "" {
		content = "菜单"
	}
	return GroupMessage{
		MessageID:    payload.ID,
		GroupOpenID:  payload.GroupOpenID,
		MemberOpenID: payload.Author.MemberOpenID,
		Content:      content,
	}, nil
}

func DecodeInteraction(data []byte) (Interaction, error) {
	var payload struct {
		ID           string `json:"id"`
		GroupOpenID  string `json:"group_openid"`
		MemberOpenID string `json:"group_member_openid"`
		Data         struct {
			Resolved struct {
				ButtonData string `json:"button_data"`
			} `json:"resolved"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return Interaction{}, err
	}
	if strings.TrimSpace(payload.ID) == "" || strings.TrimSpace(payload.GroupOpenID) == "" || strings.TrimSpace(payload.MemberOpenID) == "" || strings.TrimSpace(payload.Data.Resolved.ButtonData) == "" {
		return Interaction{}, errors.New("QQ interaction is missing identity or button data")
	}
	return Interaction{
		ID:           payload.ID,
		GroupOpenID:  payload.GroupOpenID,
		MemberOpenID: payload.MemberOpenID,
		ButtonData:   strings.TrimSpace(payload.Data.Resolved.ButtonData),
	}, nil
}
