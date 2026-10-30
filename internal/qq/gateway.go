package qq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	OpDispatch       = 0
	OpHeartbeat      = 1
	OpIdentify       = 2
	OpReconnect      = 7
	OpInvalidSession = 9
	OpHello          = 10
	OpHeartbeatAck   = 11
)

type GatewayEnvelope struct {
	ID        string          `json:"id"`
	Op        int             `json:"op"`
	Sequence  int64           `json:"s"`
	EventType string          `json:"t"`
	Data      json.RawMessage `json:"d"`
}

type identifyFrame struct {
	Op   int `json:"op"`
	Data struct {
		Token   string `json:"token"`
		Intents int64  `json:"intents"`
		Shard   [2]int `json:"shard"`
	} `json:"d"`
}

func IdentifyPayload(accessToken string, intents int64) identifyFrame {
	var frame identifyFrame
	frame.Op = OpIdentify
	frame.Data.Token = "QQBot " + accessToken
	frame.Data.Intents = intents
	frame.Data.Shard = [2]int{0, 1}
	return frame
}

func DecodeGatewayEnvelope(data []byte) (GatewayEnvelope, error) {
	var envelope GatewayEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return GatewayEnvelope{}, err
	}
	return envelope, nil
}

type Gateway struct {
	URL       string
	Token     func(context.Context) (string, error)
	Intents   int64
	Dialer    *websocket.Dialer
	OnEvent   func(context.Context, GatewayEnvelope) error
	Heartbeat time.Duration
}

func (g *Gateway) Run(ctx context.Context) error {
	if g.URL == "" {
		return errors.New("QQ gateway URL is required")
	}
	if g.Token == nil {
		return errors.New("QQ gateway token provider is required")
	}
	token, err := g.Token(ctx)
	if err != nil {
		return err
	}
	dialer := g.Dialer
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}
	conn, _, err := dialer.DialContext(ctx, g.URL, nil)
	if err != nil {
		return fmt.Errorf("connect QQ gateway: %w", err)
	}
	defer conn.Close()

	_, helloData, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read QQ gateway hello: %w", err)
	}
	hello, err := DecodeGatewayEnvelope(helloData)
	if err != nil {
		return fmt.Errorf("decode QQ gateway hello: %w", err)
	}
	if hello.Op != OpHello {
		return fmt.Errorf("expected QQ gateway hello, got op %d", hello.Op)
	}
	heartbeatInterval := g.Heartbeat
	if heartbeatInterval <= 0 {
		var payload struct {
			HeartbeatInterval int64 `json:"heartbeat_interval"`
		}
		if err := json.Unmarshal(hello.Data, &payload); err != nil {
			return fmt.Errorf("decode QQ gateway heartbeat interval: %w", err)
		}
		heartbeatInterval = time.Duration(payload.HeartbeatInterval) * time.Millisecond
	}
	if heartbeatInterval <= 0 {
		return errors.New("QQ gateway heartbeat interval is invalid")
	}

	writeMu := &sync.Mutex{}
	write := func(frame any) error {
		data, err := json.Marshal(frame)
		if err != nil {
			return err
		}
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteMessage(websocket.TextMessage, data)
	}
	if err := write(IdentifyPayload(token, g.Intents)); err != nil {
		return fmt.Errorf("identify QQ gateway: %w", err)
	}

	var sequence atomic.Int64
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				_ = write(struct {
					Op   int   `json:"op"`
					Data int64 `json:"d"`
				}{Op: OpHeartbeat, Data: sequence.Load()})
			}
		}
	}()
	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("read QQ gateway event: %w", err)
		}
		event, err := DecodeGatewayEnvelope(data)
		if err != nil {
			return fmt.Errorf("decode QQ gateway event: %w", err)
		}
		if event.Sequence > 0 {
			sequence.Store(event.Sequence)
		}
		switch event.Op {
		case OpDispatch:
			if g.OnEvent != nil {
				if err := g.OnEvent(ctx, event); err != nil {
					log.Printf("QQ event %s handler failed: %v", event.EventType, err)
				}
			}
		case OpReconnect:
			return errors.New("QQ gateway requested reconnect")
		case OpInvalidSession:
			return errors.New("QQ gateway rejected session")
		}
	}
}
