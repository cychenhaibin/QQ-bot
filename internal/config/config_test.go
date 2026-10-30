package config

import "testing"

func TestLoadFromEnvRequiresOfficialBotCredentials(t *testing.T) {
	t.Setenv("QQ_BOT_APP_ID", "")
	t.Setenv("QQ_BOT_APP_SECRET", "")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("LoadFromEnv() error = nil, want missing credentials error")
	}
}

func TestLoadFromEnvUsesSafeDefaults(t *testing.T) {
	t.Setenv("QQ_BOT_APP_ID", "app-id")
	t.Setenv("QQ_BOT_APP_SECRET", "app-secret")
	t.Setenv("QQ_BOT_API_BASE_URL", "")
	t.Setenv("NEW_API_BASE_URL", "http://new-api")
	t.Setenv("NEW_API_BOT_ID", "qq-bot")
	t.Setenv("NEW_API_SERVICE_SECRET", "service-secret")
	t.Setenv("CHARITY_NEW_API_BASE_URL", "http://charity-api")
	t.Setenv("CHARITY_NEW_API_BOT_ID", "charity-bot")
	t.Setenv("CHARITY_NEW_API_SERVICE_SECRET", "charity-secret")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.QQAPIBaseURL != "https://api.bot.qq.com" {
		t.Fatalf("QQAPIBaseURL = %q, want official default", cfg.QQAPIBaseURL)
	}
	if cfg.NewAPIBaseURL != "http://new-api" {
		t.Fatalf("NewAPIBaseURL = %q", cfg.NewAPIBaseURL)
	}
	if cfg.NewAPIBotID != "qq-bot" {
		t.Fatalf("NewAPIBotID = %q", cfg.NewAPIBotID)
	}
	if cfg.FluxA.BaseURL != "http://new-api" || cfg.Charity.BaseURL != "http://charity-api" {
		t.Fatalf("site base URLs = %+v", cfg)
	}
}

func TestLoadFromEnvRequiresNewAPIServiceCredentials(t *testing.T) {
	t.Setenv("QQ_BOT_APP_ID", "app-id")
	t.Setenv("QQ_BOT_APP_SECRET", "app-secret")
	t.Setenv("NEW_API_BASE_URL", "")
	t.Setenv("NEW_API_BOT_ID", "")
	t.Setenv("NEW_API_SERVICE_SECRET", "")
	t.Setenv("CHARITY_NEW_API_BASE_URL", "http://charity-api")
	t.Setenv("CHARITY_NEW_API_BOT_ID", "charity-bot")
	t.Setenv("CHARITY_NEW_API_SERVICE_SECRET", "charity-secret")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("LoadFromEnv() error = nil, want missing new-api credentials error")
	}
}

func TestLoadFromEnvRequiresBothSiteCredentials(t *testing.T) {
	t.Setenv("QQ_BOT_APP_ID", "app-id")
	t.Setenv("QQ_BOT_APP_SECRET", "app-secret")
	t.Setenv("FLUXA_NEW_API_BASE_URL", "http://fluxa-api")
	t.Setenv("FLUXA_NEW_API_BOT_ID", "fluxa-bot")
	t.Setenv("FLUXA_NEW_API_SERVICE_SECRET", "fluxa-secret")
	t.Setenv("CHARITY_NEW_API_BASE_URL", "")
	t.Setenv("CHARITY_NEW_API_BOT_ID", "")
	t.Setenv("CHARITY_NEW_API_SERVICE_SECRET", "")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("LoadFromEnv() error = nil, want missing Charity credentials error")
	}
}
