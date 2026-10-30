package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	QQAppID         string
	QQAppSecret     string
	QQAPIBaseURL    string
	QQTokenEndpoint string
	QQGatewayURL    string
	FluxA           SiteConfig
	Charity         SiteConfig
	// Deprecated single-site fields are retained for callers that still read them.
	NewAPIBaseURL       string
	NewAPIBotID         string
	NewAPIServiceSecret string
}

type SiteConfig struct {
	BaseURL       string
	BotID         string
	ServiceSecret string
	RechargeURL   string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		QQAppID:         strings.TrimSpace(os.Getenv("QQ_BOT_APP_ID")),
		QQAppSecret:     strings.TrimSpace(os.Getenv("QQ_BOT_APP_SECRET")),
		QQAPIBaseURL:    strings.TrimRight(strings.TrimSpace(os.Getenv("QQ_BOT_API_BASE_URL")), "/"),
		QQTokenEndpoint: strings.TrimSpace(os.Getenv("QQ_BOT_TOKEN_ENDPOINT")),
		QQGatewayURL:    strings.TrimSpace(os.Getenv("QQ_BOT_GATEWAY_URL")),
	}
	if cfg.QQAppID == "" || cfg.QQAppSecret == "" {
		return Config{}, errors.New("QQ_BOT_APP_ID and QQ_BOT_APP_SECRET are required")
	}
	if cfg.QQAPIBaseURL == "" {
		cfg.QQAPIBaseURL = "https://api.bot.qq.com"
	}
	if cfg.QQTokenEndpoint == "" {
		cfg.QQTokenEndpoint = "https://bots.qq.com/app/getAppAccessToken"
	}
	legacy := SiteConfig{
		BaseURL:       strings.TrimRight(strings.TrimSpace(os.Getenv("NEW_API_BASE_URL")), "/"),
		BotID:         strings.TrimSpace(os.Getenv("NEW_API_BOT_ID")),
		ServiceSecret: strings.TrimSpace(os.Getenv("NEW_API_SERVICE_SECRET")),
	}
	cfg.FluxA = loadSiteConfig("FLUXA", legacy, "https://catfk.com/shop/FluxA")
	cfg.Charity = loadSiteConfig("CHARITY", SiteConfig{}, "")
	cfg.NewAPIBaseURL = cfg.FluxA.BaseURL
	cfg.NewAPIBotID = cfg.FluxA.BotID
	cfg.NewAPIServiceSecret = cfg.FluxA.ServiceSecret
	if err := validateSiteConfig("FLUXA", cfg.FluxA); err != nil {
		return Config{}, err
	}
	if err := validateSiteConfig("CHARITY", cfg.Charity); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadSiteConfig(prefix string, fallback SiteConfig, defaultRechargeURL string) SiteConfig {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv(prefix+"_NEW_API_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = fallback.BaseURL
	}
	botID := strings.TrimSpace(os.Getenv(prefix + "_NEW_API_BOT_ID"))
	if botID == "" {
		botID = fallback.BotID
	}
	secret := strings.TrimSpace(os.Getenv(prefix + "_NEW_API_SERVICE_SECRET"))
	if secret == "" {
		secret = fallback.ServiceSecret
	}
	rechargeURL := strings.TrimSpace(os.Getenv(prefix + "_RECHARGE_URL"))
	if rechargeURL == "" {
		rechargeURL = defaultRechargeURL
	}
	return SiteConfig{BaseURL: baseURL, BotID: botID, ServiceSecret: secret, RechargeURL: rechargeURL}
}

func validateSiteConfig(name string, cfg SiteConfig) error {
	if cfg.BaseURL == "" || cfg.BotID == "" || cfg.ServiceSecret == "" {
		return errors.New(name + "_NEW_API_BASE_URL, " + name + "_NEW_API_BOT_ID and " + name + "_NEW_API_SERVICE_SECRET are required")
	}
	return nil
}
