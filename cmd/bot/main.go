package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/QuantumNous/qq-bot-service/internal/config"
	"github.com/QuantumNous/qq-bot-service/internal/handler"
	"github.com/QuantumNous/qq-bot-service/internal/newapi"
	"github.com/QuantumNous/qq-bot-service/internal/qq"
	"github.com/QuantumNous/qq-bot-service/internal/site"
)

type siteHandlerSet struct {
	binding  *handler.BindingHandler
	account  *handler.AccountHandler
	channel  *handler.ChannelHandler
	ai       *handler.AIHandler
	recharge handler.RechargeHandler
}

func (h siteHandlerSet) Handle(ctx context.Context, event qq.GroupMessage) (string, error) {
	if reply, err := h.binding.Handle(ctx, event); reply != "" {
		return reply, err
	}
	if reply, err := h.account.Handle(ctx, event); reply != "" {
		return reply, err
	}
	if reply, err := h.recharge.Handle(ctx, event); reply != "" {
		return reply, err
	}
	if reply, err := h.channel.Handle(ctx, event); reply != "" {
		return reply, err
	}
	return h.ai.Handle(ctx, event)
}

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	client := qq.NewClient(qq.ClientConfig{
		APIBaseURL:    cfg.QQAPIBaseURL,
		AppID:         cfg.QQAppID,
		AppSecret:     cfg.QQAppSecret,
		TokenEndpoint: cfg.QQTokenEndpoint,
	})
	gatewayURL := cfg.QQGatewayURL
	if gatewayURL == "" {
		gatewayURL, err = client.GetGatewayURL(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}
	menuHandler := handler.MenuHandler{Sender: client}
	clients := map[site.ID]siteHandlerSet{}
	for id, siteConfig := range map[site.ID]config.SiteConfig{
		site.FluxA:   cfg.FluxA,
		site.Charity: cfg.Charity,
	} {
		newAPIClient := newapi.NewClient(newapi.ClientConfig{
			BaseURL: siteConfig.BaseURL,
			BotID:   siteConfig.BotID,
			Secret:  siteConfig.ServiceSecret,
		})
		clients[id] = siteHandlerSet{
			binding:  handler.NewBindingHandler(newAPIClient),
			account:  handler.NewAccountHandler(newAPIClient),
			channel:  handler.NewChannelHandler(newAPIClient),
			ai:       handler.NewAIHandler(newAPIClient),
			recharge: handler.NewRechargeHandler(siteConfig.RechargeURL),
		}
	}
	siteRouter := handler.NewSiteRouter()
	gateway := qq.Gateway{
		URL:     gatewayURL,
		Intents: (1 << 25) | (1 << 26),
		Token:   client.AccessToken,
		OnEvent: func(ctx context.Context, envelope qq.GatewayEnvelope) error {
			dispatch := func(selected site.ID, routedEvent qq.GroupMessage) (string, error) {
				selectedHandlers, ok := clients[selected]
				if !ok {
					return "站点暂时不可用，请重新选择站点。", nil
				}
				return selectedHandlers.Handle(ctx, routedEvent)
			}
			process := func(event qq.GroupMessage, interactionID string) error {
				reply, handled, routeErr := siteRouter.Handle(event, dispatch)
				if handled {
					if routeErr != nil {
						log.Printf("site-routed request failed: %v", routeErr)
					}
					if reply != "" {
						return handler.SendGroupReply(ctx, client, event.GroupOpenID, reply, event.MessageID, interactionID, event.MemberOpenID)
					}
					return nil
				}
				if interactionID != "" {
					if handler.IsMenuCommand(event.Content) {
						return handler.SendGroupReply(ctx, client, event.GroupOpenID, handler.BuildMainMenu(), "", interactionID, event.MemberOpenID)
					}
					return nil
				}
				return menuHandler.Handle(ctx, event)
			}
			switch envelope.EventType {
			case "GROUP_AT_MESSAGE_CREATE", "GROUP_MESSAGE_CREATE":
				event, err := qq.DecodeGroupMessage(envelope.Data, envelope.EventType)
				if err != nil {
					return err
				}
				return process(event, "")
			case "INTERACTION_CREATE":
				interaction, err := qq.DecodeInteraction(envelope.Data)
				if err != nil {
					return err
				}
				if err := client.AcknowledgeInteraction(ctx, interaction.ID); err != nil {
					return err
				}
				if envelope.ID == "" {
					return fmt.Errorf("QQ interaction is missing gateway event id")
				}
				return process(qq.GroupMessage{
					GroupOpenID:  interaction.GroupOpenID,
					MemberOpenID: interaction.MemberOpenID,
					Content:      interaction.ButtonData,
				}, envelope.ID)
			default:
				return nil
			}
		},
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := gateway.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
