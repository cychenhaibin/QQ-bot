package qq

import (
	"encoding/json"
	"testing"
)

func TestIdentifyPayloadUsesOfficialQQBotTokenFormat(t *testing.T) {
	payload := IdentifyPayload("access-token", 1<<25)
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	want := `{"op":2,"d":{"token":"QQBot access-token","intents":33554432,"shard":[0,1]}}`
	if string(data) != want {
		t.Fatalf("identify payload = %s, want %s", data, want)
	}
}

func TestDecodeGatewayEnvelopePreservesSequenceAndEventData(t *testing.T) {
	envelope, err := DecodeGatewayEnvelope([]byte(`{"id":"event-1","op":0,"s":7,"t":"GROUP_AT_MESSAGE_CREATE","d":{"id":"message-1"}}`))
	if err != nil {
		t.Fatalf("DecodeGatewayEnvelope() error = %v", err)
	}
	if envelope.Op != OpDispatch || envelope.Sequence != 7 || envelope.EventType != "GROUP_AT_MESSAGE_CREATE" {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
	if string(envelope.Data) != `{"id":"message-1"}` {
		t.Fatalf("Data = %s", envelope.Data)
	}
	if envelope.ID != "event-1" {
		t.Fatalf("gateway event ID = %q, want event-1", envelope.ID)
	}
}
