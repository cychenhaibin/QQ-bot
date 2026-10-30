package handler

import "testing"

func TestKeyboardForPromptBuildsMainMenuButtons(t *testing.T) {
	keyboard, ok := KeyboardForPrompt(BuildMainMenu())
	if !ok || len(keyboard.Content.Rows) != 3 {
		t.Fatalf("KeyboardForPrompt(main menu) = %#v, %v", keyboard, ok)
	}
	if got := keyboard.Content.Rows[0].Buttons[0].Action.Data; got != "每日签到" {
		t.Fatalf("first button data = %q", got)
	}
}

func TestKeyboardForPromptBuildsSiteAndReplacementButtons(t *testing.T) {
	siteKeyboard, ok := KeyboardForPrompt(siteSelectionPrompt())
	if !ok || len(siteKeyboard.Content.Rows) != 1 || len(siteKeyboard.Content.Rows[0].Buttons) != 2 {
		t.Fatalf("site keyboard = %#v, %v", siteKeyboard, ok)
	}
	replaceKeyboard, ok := KeyboardForPrompt("检测到账号已绑定非 QQ 邮箱（example.com），如果不换绑，后续无法使用 Bot。是否解绑后重新绑定？请回复：是/否")
	if !ok || len(replaceKeyboard.Content.Rows[0].Buttons) != 2 {
		t.Fatalf("replacement keyboard = %#v, %v", replaceKeyboard, ok)
	}
}

func TestKeyboardForPromptBuildsDynamicAIButtons(t *testing.T) {
	groupKeyboard, ok := KeyboardForPrompt("请选择使用哪个分组：free、pro\n未在 API 密钥中创建的分组将自动创建。发送“停”将停止本轮对话。")
	if !ok || len(groupKeyboard.Content.Rows[0].Buttons) != 3 {
		t.Fatalf("group keyboard = %#v, %v", groupKeyboard, ok)
	}
	modelKeyboard, ok := KeyboardForPrompt("pro分组下有以下模型请先选择：gpt-4、gpt-5")
	if !ok || len(modelKeyboard.Content.Rows[0].Buttons) != 2 {
		t.Fatalf("model keyboard = %#v, %v", modelKeyboard, ok)
	}
}

func TestKeyboardForPromptIgnoresFreeFormReplies(t *testing.T) {
	if _, ok := KeyboardForPrompt("请输入要绑定的邮箱地址，例如：xxxxx@qq.com"); ok {
		t.Fatal("free-form email prompt unexpectedly produced buttons")
	}
}
