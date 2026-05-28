package main

import (
	"context"
	"fmt"
	"os"
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

	// --- Fyne GUI ---
	a := fyneApp.NewWithID("com.wispr-vibe.app")
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("Wispr Vibe — Speech to Text")
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	// UI widgets
	statusLabel := widget.NewLabel("🟢 Ready — press Ctrl+Shift+R or click Record")
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
	toggleRecording := func() {
		if !recording {
			if err := application.StartRecording(); err != nil {
				statusLabel.SetText(fmt.Sprintf("❌ %v", err))
				return
			}
			recording = true
			recordStart = time.Now()
			recordBtn.SetText("⏹ Stop & Transcribe")
			statusLabel.SetText("🎙 Recording... (Ctrl+Shift+R to stop)")
			resultLabel.SetText("")

			timerDone = make(chan struct{})
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-timerDone:
						return
					case <-ticker.C:
						elapsed := time.Since(recordStart).Truncate(100 * time.Millisecond)
						timerLabel.SetText(fmt.Sprintf("⏱ %s", elapsed))
					}
				}
			}()
		} else {
			recording = false
			close(timerDone)
			recordBtn.SetText("⏳ Transcribing...")
			recordBtn.Disable()
			statusLabel.SetText("⏳ Processing audio...")

			go func() {
				// Minimize window so the PREVIOUS app regains focus
				w.Hide()
				time.Sleep(300 * time.Millisecond)

				text, err := application.StopAndProcess(ctx)
				if err != nil {
					w.Show()
					statusLabel.SetText(fmt.Sprintf("❌ %v", err))
				} else {
					// Text was already typed into active window by app.StopAndProcess
					w.Show()
					statusLabel.SetText("✅ Text inserted at cursor!")
					resultLabel.SetText(text)
				}
				recordBtn.SetText("🎙 Record")
				recordBtn.Enable()
				timerLabel.SetText("")
			}()
		}
	}

	recordBtn.OnTapped = toggleRecording

	// Register global hotkey: Ctrl+Shift+R (VK_R = 0x52)
	hk, err := hotkey.Register(1, hotkey.ModControl|hotkey.ModShift, 0x52, toggleRecording)
	if err != nil {
		log.Warn("global hotkey registration failed", "error", err)
	} else {
		log.Info("global hotkey registered: Ctrl+Shift+R")
	}

	// Settings button
	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		showSettings(a, cfg)
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

	hotkeyHint := widget.NewLabel("Global hotkey: Ctrl+Shift+R (works from any app)")
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
				a.Quit()
			}),
		))
	}

	w.ShowAndRun()

	// Cleanup
	if hk != nil {
		hk.Unregister()
	}
}

func showSettings(a fyne.App, cfg *config.Config) {
	w := a.NewWindow("Settings")
	w.Resize(fyne.NewSize(380, 280))

	engineText := cfg.STTEngine
	if engineText == "whisper_local" {
		engineText = "Local Whisper (offline)"
	}

	info := container.NewVBox(
		widget.NewLabelWithStyle("Configuration", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Engine: %s", engineText)),
		widget.NewLabel(fmt.Sprintf("Provider: %s", orDefault(cfg.Provider, "local"))),
		widget.NewLabel(fmt.Sprintf("Language: %s", orDefault(cfg.Language, "auto-detect"))),
		widget.NewLabel(fmt.Sprintf("Model: %s", orDefault(cfg.WhisperModel, "base"))),
		widget.NewSeparator(),
		widget.NewLabel("Hotkey: Ctrl+Shift+R (toggle record)"),
		widget.NewSeparator(),
		widget.NewLabel("To reconfigure: delete ~/.wispr-vibe/config.json"),
		widget.NewLabel("and restart the app."),
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
