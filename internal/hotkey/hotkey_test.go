package hotkey

import (
	"testing"
)

func TestParseHotkeyValid(t *testing.T) {
	tests := []struct {
		input    string
		wantMods uint32
		wantVK   uint32
	}{
		{"Ctrl+Shift+R", ModControl | ModShift, 'R'},
		{"Ctrl+R", ModControl, 'R'},
		{"Alt+Z", ModAlt, 'Z'},
		{"Ctrl+Alt+Delete", ModControl | ModAlt, 0x2E},
		{"Ctrl+Shift+F1", ModControl | ModShift, 0x70},
		{"Ctrl+Shift+F12", ModControl | ModShift, 0x7B},
		{"Win+Space", ModWin, 0x20},
		{"Shift+1", ModShift, '1'},
		{"Ctrl+0", ModControl, '0'},
		{"Alt+Shift+A", ModAlt | ModShift, 'A'},
		{"Ctrl+Alt+Shift+Win+F5", ModControl | ModAlt | ModShift | ModWin, 0x74},
		{"ctrl+shift+r", ModControl | ModShift, 'R'},
		{"CTRL+SHIFT+R", ModControl | ModShift, 'R'},
		{"Control+R", ModControl, 'R'},
		{"Super+A", ModWin, 'A'},
		{"Ctrl+Enter", ModControl, 0x0D},
		{"Ctrl+Escape", ModControl, 0x1B},
		{"Ctrl+Tab", ModControl, 0x09},
		{"Alt+F4", ModAlt, 0x73},
		{"Ctrl+Home", ModControl, 0x24},
		{"Ctrl+End", ModControl, 0x23},
		{"Ctrl+PageUp", ModControl, 0x21},
		{"Ctrl+PageDown", ModControl, 0x22},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mods, vk, err := ParseHotkey(tt.input)
			if err != nil {
				t.Fatalf("ParseHotkey(%q) error: %v", tt.input, err)
			}
			if mods != tt.wantMods {
				t.Errorf("mods = 0x%04X, want 0x%04X", mods, tt.wantMods)
			}
			if vk != tt.wantVK {
				t.Errorf("vk = 0x%04X, want 0x%04X", vk, tt.wantVK)
			}
		})
	}
}

func TestParseHotkeyInvalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty string"},
		{"Ctrl+", "no key after modifier"},
		{"Ctrl+Shift", "only modifiers, no key"},
		{"Ctrl+Shift+UnknownKey", "unknown key name"},
		{"Ctrl+Shift+😀", "emoji as key"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, _, err := ParseHotkey(tt.input)
			if err == nil {
				t.Errorf("ParseHotkey(%q) should error: %s", tt.input, tt.desc)
			}
		})
	}
}

func TestFormatHotkey(t *testing.T) {
	tests := []struct {
		mods   uint32
		vk     uint32
		expect string
	}{
		{ModControl | ModShift, 'R', "Ctrl+Shift+R"},
		{ModControl, 'A', "Ctrl+A"},
		{ModAlt, 0x70, "Alt+F1"},
		{ModControl | ModAlt | ModShift, 0x2E, "Ctrl+Alt+Shift+Delete"},
		{ModWin, 0x20, "Win+Space"},
		{0, 'X', "X"},
		{ModControl | ModShift, '5', "Ctrl+Shift+5"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			got := FormatHotkey(tt.mods, tt.vk)
			if got != tt.expect {
				t.Errorf("FormatHotkey(0x%X, 0x%X) = %q, want %q", tt.mods, tt.vk, got, tt.expect)
			}
		})
	}
}

func TestParseFormatRoundtrip(t *testing.T) {
	combos := []string{
		"Ctrl+Shift+R",
		"Ctrl+A",
		"Alt+F1",
		"Win+Space",
		"Ctrl+Shift+5",
	}

	for _, combo := range combos {
		t.Run(combo, func(t *testing.T) {
			mods, vk, err := ParseHotkey(combo)
			if err != nil {
				t.Fatalf("ParseHotkey(%q) error: %v", combo, err)
			}
			formatted := FormatHotkey(mods, vk)
			if formatted != combo {
				t.Errorf("roundtrip: %q → FormatHotkey → %q", combo, formatted)
			}
		})
	}
}

func TestKeyNameToVKLetters(t *testing.T) {
	for c := byte('A'); c <= 'Z'; c++ {
		vk, ok := keyNameToVK(string(c))
		if !ok {
			t.Errorf("keyNameToVK(%q) not found", string(c))
		}
		if vk != uint32(c) {
			t.Errorf("keyNameToVK(%q) = 0x%X, want 0x%X", string(c), vk, c)
		}
	}
}

func TestKeyNameToVKDigits(t *testing.T) {
	for c := byte('0'); c <= '9'; c++ {
		vk, ok := keyNameToVK(string(c))
		if !ok {
			t.Errorf("keyNameToVK(%q) not found", string(c))
		}
		if vk != uint32(c) {
			t.Errorf("keyNameToVK(%q) = 0x%X, want 0x%X", string(c), vk, c)
		}
	}
}

func TestKeyNameToVKCaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  uint32
	}{
		{"r", 'R'},
		{"a", 'A'},
		{"z", 'Z'},
		{"f1", 0x70},
		{"space", 0x20},
		{"SPACE", 0x20},
		{"Space", 0x20},
		{"enter", 0x0D},
		{"ENTER", 0x0D},
		{"return", 0x0D},
		{"esc", 0x1B},
		{"escape", 0x1B},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			vk, ok := keyNameToVK(tt.input)
			if !ok {
				t.Fatalf("keyNameToVK(%q) not found", tt.input)
			}
			if vk != tt.want {
				t.Errorf("keyNameToVK(%q) = 0x%X, want 0x%X", tt.input, vk, tt.want)
			}
		})
	}
}

func TestVkToKeyNameSpecialKeys(t *testing.T) {
	tests := []struct {
		vk   uint32
		want string
	}{
		{0x70, "F1"},
		{0x7B, "F12"},
		{0x20, "Space"},
		{0x0D, "Enter"},
		{0x09, "Tab"},
		{0x1B, "Esc"},
		{0x08, "Backspace"},
		{0x2E, "Delete"},
		{0x2D, "Insert"},
		{0x24, "Home"},
		{0x23, "End"},
		{0x21, "PageUp"},
		{0x22, "PageDown"},
		{0x26, "Up"},
		{0x28, "Down"},
		{0x25, "Left"},
		{0x27, "Right"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := vkToKeyName(tt.vk)
			if got != tt.want {
				t.Errorf("vkToKeyName(0x%X) = %q, want %q", tt.vk, got, tt.want)
			}
		})
	}
}
