package qq

import "testing"

func TestDecodeGroupAtMessage(t *testing.T) {
	data := []byte(`{"id":"message-1","group_openid":"group-1","content":" /菜单 ","author":{"member_openid":"member-1"}}`)

	event, err := DecodeGroupMessage(data, "GROUP_AT_MESSAGE_CREATE")
	if err != nil {
		t.Fatalf("DecodeGroupMessage() error = %v", err)
	}
	if event.GroupOpenID != "group-1" || event.MemberOpenID != "member-1" {
		t.Fatalf("unexpected identity: %+v", event)
	}
	if event.Content != "/菜单" {
		t.Fatalf("Content = %q, want trimmed command", event.Content)
	}
}

func TestDecodeGroupMessageRejectsOtherEventTypes(t *testing.T) {
	if _, err := DecodeGroupMessage([]byte(`{}`), "C2C_MESSAGE_CREATE"); err == nil {
		t.Fatal("DecodeGroupMessage() error = nil, want unsupported event error")
	}
}

func TestDecodeInteraction(t *testing.T) {
	data := []byte(`{"id":"interaction-1","type":11,"chat_type":1,"group_openid":"group-1","group_member_openid":"member-1","data":{"resolved":{"button_data":"FluxA","button_id":"site_fluxa"}}}`)

	interaction, err := DecodeInteraction(data)
	if err != nil {
		t.Fatalf("DecodeInteraction() error = %v", err)
	}
	if interaction.ID != "interaction-1" || interaction.GroupOpenID != "group-1" || interaction.MemberOpenID != "member-1" || interaction.ButtonData != "FluxA" {
		t.Fatalf("unexpected interaction: %+v", interaction)
	}
}

func TestDecodeInteractionRejectsMissingIdentity(t *testing.T) {
	if _, err := DecodeInteraction([]byte(`{"id":"interaction-1","data":{"resolved":{"button_data":"FluxA"}}}`)); err == nil {
		t.Fatal("DecodeInteraction() error = nil, want missing identity error")
	}
}
