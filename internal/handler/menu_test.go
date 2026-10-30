package handler

import (
	"strings"
	"testing"
)

func TestBuildMainMenuListsSupportedFeatures(t *testing.T) {
	menu := BuildMainMenu()
	for _, item := range []string{"每日签到", "查询余额", "绑定账号", "AI 对话", "充值中心", "模型渠道"} {
		if !strings.Contains(menu, item) {
			t.Fatalf("menu does not contain %q: %s", item, menu)
		}
	}
}
