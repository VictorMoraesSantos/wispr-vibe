package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/victorlui/wispr-vibe/internal/app"
	"github.com/victorlui/wispr-vibe/internal/config"
	"github.com/victorlui/wispr-vibe/internal/hotkey"
	"github.com/victorlui/wispr-vibe/internal/logger"
	"github.com/victorlui/wispr-vibe/internal/toast"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load: %v\n", err)
		os.Exit(1)
	}

	if cfg.NeedsSetup() {
		if err := config.RunSetup(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "setup: %v\n", err)
			os.Exit(1)
		}
	}

	logger.Init(cfg.LogLevel)
	log := logger.Get()

	application := app.New(cfg, log)
	ctx := context.Background()

	// Create native toast overlay (bottom of screen, always on top)
	toastOverlay := toast.New()

	// --- Fyne GUI ---
	a := fyneApp.NewWithID("com.wispr-vibe.app")
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Wispr Vibe — Speech to Text")
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	// Hotkey combo from config
	hotkeyCombo := cfg.Hotkey
	if hotkeyCombo == "" {
		hotkeyCombo = "Ctrl+Shift+R"
	}

	// UI widgets
	statusLabel := widget.NewLabel(fmt.Sprintf("🟢 Ready — press %s or click Record", hotkeyCombo))
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	timerLabel := widget.NewLabel("")
	timerLabel.Alignment = fyne.TextAlignCenter

	var recording bool
	var recordStart time.Time
	var timerDone chan struct{}

	recordBtn := widget.NewButton("🎙 Record", nil)
	recordBtn.Importance = widget.HighImportance

	// Toggle recording function (shared by button and hotkey)
	// Must wrap all UI calls in fyne.Do when called from hotkey background thread
	toggleRecording := func() {
		fyne.Do(func() {
			if !recording {
				if err := application.StartRecording(); err != nil {
					statusLabel.SetText(fmt.Sprintf("❌ %v", err))
					return
				}
				recording = true
				recordStart = time.Now()
				recordBtn.SetText("⏹ Stop & Transcribe")
				statusLabel.SetText(fmt.Sprintf("🎙 Recording... (%s to stop)", hotkeyCombo))
				resultLabel.SetText("")

				// Show toast overlay at bottom of screen
				toastOverlay.Show("🎙 Recording...")

				timerDone = make(chan struct{})
				go func() {
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-timerDone:
							return
						case <-ticker.C:
							secs := int(time.Since(recordStart).Seconds())
							toastOverlay.SetText(toast.FormatRecordingText(secs))
							fyne.Do(func() {
								elapsed := time.Since(recordStart).Truncate(time.Second)
								timerLabel.SetText(fmt.Sprintf("⏱ %s", elapsed))
							})
						}
					}
				}()
			} else {
				recording = false
				close(timerDone)
				recordBtn.SetText("⏳ Transcribing...")
				recordBtn.Disable()
				statusLabel.SetText("⏳ Processing audio...")

				// Update toast to show processing
				toastOverlay.SetText("⏳ Transcribing...")

				go func() {
					// Hide main window so previous app regains focus for paste
					fyne.Do(func() {
						w.Hide()
					})
					time.Sleep(300 * time.Millisecond)

					text, err := application.StopAndProcess(ctx)

					// Hide toast
					toastOverlay.Hide()

					fyne.Do(func() {
						if err != nil {
							w.Show()
							statusLabel.SetText(fmt.Sprintf("❌ %v", err))
						} else {
							w.Show()
							statusLabel.SetText("✅ Text inserted at cursor!")
							resultLabel.SetText(text)
						}
						recordBtn.SetText("🎙 Record")
						recordBtn.Enable()
						timerLabel.SetText("")
					})
				}()
			}
		})
	}

	recordBtn.OnTapped = toggleRecording

	// Register global hotkey from config
	hk, err := hotkey.RegisterFromString(1, hotkeyCombo, toggleRecording)
	if err != nil {
		log.Warn("global hotkey registration failed", "error", err, "hotkey", hotkeyCombo)
	} else {
		log.Info("global hotkey registered", "hotkey", hotkeyCombo)
	}

	// Settings button
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		showSettings(a, cfg, w, func() {
			// Re-register hotkey after settings change
			if hk != nil {
				hk.Unregister()
			}
			newCombo := cfg.Hotkey
			if newCombo == "" {
				newCombo = "Ctrl+Shift+R"
			}
			newHk, err := hotkey.RegisterFromString(1, newCombo, toggleRecording)
			if err != nil {
				log.Warn("hotkey re-register failed", "error", err)
			} else {
				hk = newHk
				log.Info("hotkey updated", "hotkey", newCombo)
			}
		})
	})

	// Minimize button
	minimizeBtn := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		w.Hide()
	})

	// Layout
	header := container.NewHBox(
		widget.NewLabelWithStyle("Wispr Vibe", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		minimizeBtn,
		settingsBtn,
	)

	hotkeyHint := widget.NewLabel(fmt.Sprintf("Hotkey: %s (works from any app)", hotkeyCombo))
	hotkeyHint.Alignment = fyne.TextAlignCenter
	hotkeyHint.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		container.NewCenter(statusLabel),
		container.NewCenter(timerLabel),
		layout.NewSpacer(),
		container.NewCenter(recordBtn),
		container.NewCenter(hotkeyHint),
		layout.NewSpacer(),
		widget.NewSeparator(),
		resultLabel,
	)

	w.SetContent(content)

	// Close hides to tray instead of quitting
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	// System tray
	if desk, ok := a.(interface {
		SetSystemTrayMenu(menu *fyne.Menu)
	}); ok {
		desk.SetSystemTrayMenu(fyne.NewMenu("Wispr Vibe",
			fyne.NewMenuItem("Show", func() { w.Show() }),
			fyne.NewMenuItem("Quit", func() {
				if hk != nil {
					hk.Unregister()
				}
				toastOverlay.Destroy()
				a.Quit()
			}),
		))
	}

	w.ShowAndRun()

	// Cleanup
	if hk != nil {
		hk.Unregister()
	}
	toastOverlay.Destroy()
}

func showSettings(a fyne.App, cfg *config.Config, parent fyne.Window, onHotkeyChanged func()) {
	w := a.NewWindow("Settings")
	w.Resize(fyne.NewSize(420, 350))

	engineText := cfg.STTEngine
	if engineText == "whisper_local" {
		engineText = "Local Whisper (offline)"
	}

	currentHotkey := cfg.Hotkey
	if currentHotkey == "" {
		currentHotkey = "Ctrl+Shift+R"
	}

	// Hotkey entry
	hotkeyEntry := widget.NewEntry()
	hotkeyEntry.SetText(currentHotkey)
	hotkeyEntry.SetPlaceHolder("Ex: Ctrl+Shift+R, Alt+Z, Ctrl+F9")

	hotkeyHelp := widget.NewLabel("Format: Modifier+Key (Ctrl, Shift, Alt, Win + any key)")
	hotkeyHelp.TextStyle = fyne.TextStyle{Italic: true}
	hotkeyHelp.Wrapping = fyne.TextWrapWord

	saveStatus := widget.NewLabel("")

	saveBtn := widget.NewButton("💾 Save Hotkey", func() {
		newCombo := strings.TrimSpace(hotkeyEntry.Text)
		if newCombo == "" {
			saveStatus.SetText("❌ Hotkey cannot be empty")
			return
		}

		// Validate
		_, _, err := hotkey.ParseHotkey(newCombo)
		if err != nil {
			saveStatus.SetText(fmt.Sprintf("❌ Invalid: %v", err))
			return
		}

		cfg.Hotkey = newCombo
		if err := config.Save(cfg, ""); err != nil {
			saveStatus.SetText(fmt.Sprintf("❌ Save failed: %v", err))
			return
		}

		saveStatus.SetText(fmt.Sprintf("✅ Hotkey saved: %s", newCombo))

		// Notify to re-register
		if onHotkeyChanged != nil {
			onHotkeyChanged()
		}
	})
	saveBtn.Importance = widget.HighImportance

	info := container.NewVBox(
		widget.NewLabelWithStyle("⚙️ Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Engine: %s", engineText)),
		widget.NewLabel(fmt.Sprintf("Provider: %s", orDefault(cfg.Provider, "local"))),
		widget.NewLabel(fmt.Sprintf("Language: %s", orDefault(cfg.Language, "auto-detect"))),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("🎹 Push-to-Talk Hotkey", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		hotkeyHelp,
		hotkeyEntry,
		saveBtn,
		saveStatus,
		widget.NewSeparator(),
		widget.NewLabel("Press hotkey once → start recording"),
		widget.NewLabel("Press again → stop & transcribe to cursor"),
	)

	w.SetContent(container.NewPadded(info))
	w.CenterOnScreen()
	w.Show()
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
