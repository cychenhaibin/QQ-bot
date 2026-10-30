package handler

import (
	"strconv"
	"strings"

	"github.com/QuantumNous/qq-bot-service/internal/qq"
)

func KeyboardForPrompt(prompt string) (qq.Keyboard, bool) {
	prompt = strings.TrimSpace(prompt)
	switch {
	case prompt == BuildMainMenu():
		return newKeyboard([][]string{
			{"每日签到", "查询余额"},
			{"绑定账号", "AI 对话"},
			{"充值中心", "模型渠道"},
		}), true
	case prompt == siteSelectionPrompt():
		return newKeyboard([][]string{{"FluxA", "Charity"}}), true
	case strings.Contains(prompt, "是否解绑后重新绑定") || prompt == "请回复“是”或“否”。":
		return newKeyboard([][]string{{"是", "否"}}), true
	case strings.HasPrefix(prompt, "请选择使用哪个分组："):
		choices := choicesAfter(prompt, "请选择使用哪个分组：")
		choices = append(choices, "停")
		return newChoiceKeyboard(choices), len(choices) > 1
	case strings.Contains(prompt, "分组下有以下模型请先选择："):
		choices := choicesAfter(prompt, "模型请先选择：")
		return newChoiceKeyboard(choices), len(choices) > 0
	default:
		return qq.Keyboard{}, false
	}
}

func newChoiceKeyboard(choices []string) qq.Keyboard {
	rows := make([][]string, 0, (len(choices)+4)/5)
	for len(choices) > 0 {
		size := min(5, len(choices))
		rows = append(rows, choices[:size])
		choices = choices[size:]
	}
	return newKeyboard(rows)
}

func newKeyboard(rows [][]string) qq.Keyboard {
	keyboardRows := make([]qq.KeyboardRow, 0, len(rows))
	buttonIndex := 0
	for _, row := range rows {
		buttons := make([]qq.KeyboardButton, 0, len(row))
		for _, label := range row {
			label = strings.TrimSpace(label)
			if label == "" {
				continue
			}
			buttons = append(buttons, qq.KeyboardButton{
				ID:         "choice_" + strconv.Itoa(buttonIndex),
				RenderData: qq.RenderData{Label: label, VisitedLabel: label, Style: 1},
				Action: qq.ButtonAction{
					Type:          1,
					Permission:    qq.ButtonPermission{Type: 2},
					Data:          label,
					UnsupportTips: "请发送按钮上的文字",
				},
			})
			buttonIndex++
		}
		if len(buttons) > 0 {
			keyboardRows = append(keyboardRows, qq.KeyboardRow{Buttons: buttons})
		}
	}
	return qq.Keyboard{Content: qq.KeyboardContent{Rows: keyboardRows}}
}

func choicesAfter(prompt, marker string) []string {
	_, value, _ := strings.Cut(prompt, marker)
	value = strings.TrimSpace(value)
	if newline := strings.IndexByte(value, '\n'); newline >= 0 {
		value = value[:newline]
	}
	parts := strings.Split(value, "、")
	choices := make([]string, 0, len(parts))
	for _, part := range parts {
		if choice := strings.TrimSpace(part); choice != "" {
			choices = append(choices, choice)
		}
	}
	return choices
}

// KeyboardPrompt removes choices already rendered by the keyboard.
func KeyboardPrompt(prompt string) string {
	switch {
	case prompt == BuildMainMenu():
		return "请选择功能："
	case prompt == siteSelectionPrompt():
		return "请选择要使用的站点："
	case strings.HasPrefix(prompt, "请选择使用哪个分组："):
		_, suffix, _ := strings.Cut(prompt, "\n")
		return "请选择使用哪个分组：\n" + suffix
	case strings.Contains(prompt, "分组下有以下模型请先选择："):
		prefix, _, _ := strings.Cut(prompt, "模型请先选择：")
		return prefix + "模型，请点击选择："
	default:
		return prompt
	}
}
