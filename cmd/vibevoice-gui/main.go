package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
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
	toastOverlay := toast.New()

	a := fyneApp.NewWithID("com.wispr-vibe.app")
	a.Settings().SetTheme(newVibeTheme())

	w := a.NewWindow("Wispr Vibe")
	w.Resize(fyne.NewSize(380, 340))
	w.CenterOnScreen()

	hotkeyCombo := cfg.Hotkey
	if hotkeyCombo == "" {
		hotkeyCombo = "Ctrl+Shift+R"
	}

	statusIndicator := canvas.NewCircle(colorReady)
	statusIndicator.Resize(fyne.NewSize(10, 10))

	statusLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Ready  ·  %s", hotkeyCombo),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	timerLabel := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})

	var recording bool
	var recordStart time.Time
	var timerDone chan struct{}

	recordBtn := widget.NewButton("Record", nil)
	recordBtn.Importance = widget.HighImportance

	toggleRecording := func() {
		fyne.Do(func() {
			if !recording {
				if err := application.StartRecording(); err != nil {
					statusLabel.SetText(fmt.Sprintf("Error: %v", err))
					statusIndicator.FillColor = colorError
					statusIndicator.Refresh()
					return
				}
				recording = true
				recordStart = time.Now()
				recordBtn.SetText("Stop & Transcribe")
				statusLabel.SetText(fmt.Sprintf("Recording  ·  %s to stop", hotkeyCombo))
				statusIndicator.FillColor = colorRecording
				statusIndicator.Refresh()
				resultLabel.SetText("")

				toastOverlay.Show("Recording...")

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
								timerLabel.SetText(elapsed.String())
							})
						}
					}
				}()
			} else {
				recording = false
				close(timerDone)
				recordBtn.SetText("Transcribing...")
				recordBtn.Disable()
				statusLabel.SetText("Processing audio...")
				statusIndicator.FillColor = colorProcessing
				statusIndicator.Refresh()
				toastOverlay.SetText("Transcribing...")

				go func() {
					fyne.Do(func() { w.Hide() })
					time.Sleep(300 * time.Millisecond)

					text, err := application.StopAndProcess(ctx)
					toastOverlay.Hide()

					fyne.Do(func() {
						if err != nil {
							statusLabel.SetText(fmt.Sprintf("Error: %v", err))
							statusIndicator.FillColor = colorError
						} else {
							statusLabel.SetText("Done  ·  text inserted at cursor")
							statusIndicator.FillColor = colorSuccess
							resultLabel.SetText(text)
						}
						statusIndicator.Refresh()
						recordBtn.SetText("Record")
						recordBtn.Enable()
						timerLabel.SetText("")
					})
				}()
			}
		})
	}

	recordBtn.OnTapped = toggleRecording

	hk, err := hotkey.RegisterFromString(1, hotkeyCombo, toggleRecording)
	if err != nil {
		log.Warn("hotkey registration failed", "error", err)
	}

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		showSettings(a, cfg, w, func() {
			if hk != nil {
				hk.Unregister()
			}
			newCombo := cfg.Hotkey
			if newCombo == "" {
				newCombo = "Ctrl+Shift+R"
			}
			newHk, err := hotkey.RegisterFromString(1, newCombo, toggleRecording)
			if err == nil {
				hk = newHk
			}
		})
	})

	minimizeBtn := widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		w.Hide()
	})

	titleLabel := widget.NewLabelWithStyle("Wispr Vibe", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	versionLabel := widget.NewLabelWithStyle("v1.0", fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	versionLabel.Importance = widget.LowImportance

	header := container.NewHBox(
		titleLabel,
		versionLabel,
		layout.NewSpacer(),
		minimizeBtn,
		settingsBtn,
	)

	statusRow := container.NewHBox(
		layout.NewSpacer(),
		container.NewPadded(statusIndicator),
		statusLabel,
		layout.NewSpacer(),
	)

	hotkeyHint := widget.NewLabelWithStyle(
		fmt.Sprintf("%s from any app", hotkeyCombo),
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	hotkeyHint.Importance = widget.LowImportance

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		spacer(12),
		statusRow,
		container.NewCenter(timerLabel),
		spacer(8),
		layout.NewSpacer(),
		container.NewCenter(recordBtn),
		spacer(4),
		container.NewCenter(hotkeyHint),
		layout.NewSpacer(),
		widget.NewSeparator(),
		resultLabel,
	)

	w.SetContent(container.NewPadded(content))
	w.SetCloseIntercept(func() { w.Hide() })

	if desk, ok := a.(interface{ SetSystemTrayMenu(menu *fyne.Menu) }); ok {
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

	if hk != nil {
		hk.Unregister()
	}
	toastOverlay.Destroy()
}

func showSettings(a fyne.App, cfg *config.Config, parent fyne.Window, onHotkeyChanged func()) {
	w := a.NewWindow("Settings")
	w.Resize(fyne.NewSize(400, 380))

	engineText := cfg.STTEngine
	if engineText == "whisper_local" {
		engineText = "Local Whisper (offline)"
	}

	currentHotkey := cfg.Hotkey
	if currentHotkey == "" {
		currentHotkey = "Ctrl+Shift+R"
	}

	hotkeyEntry := widget.NewEntry()
	hotkeyEntry.SetText(currentHotkey)
	hotkeyEntry.SetPlaceHolder("Ctrl+Shift+R, Alt+Z, Ctrl+F9...")

	saveStatus := widget.NewLabel("")

	saveBtn := widget.NewButton("Save", func() {
		newCombo := strings.TrimSpace(hotkeyEntry.Text)
		if newCombo == "" {
			saveStatus.SetText("Hotkey cannot be empty")
			return
		}

		if _, _, err := hotkey.ParseHotkey(newCombo); err != nil {
			saveStatus.SetText(fmt.Sprintf("Invalid: %v", err))
			return
		}

		cfg.Hotkey = newCombo
		if err := config.Save(cfg, ""); err != nil {
			saveStatus.SetText(fmt.Sprintf("Save failed: %v", err))
			return
		}

		saveStatus.SetText(fmt.Sprintf("Saved: %s", newCombo))
		if onHotkeyChanged != nil {
			onHotkeyChanged()
		}
	})
	saveBtn.Importance = widget.HighImportance

	engineSection := container.NewVBox(
		widget.NewLabelWithStyle("Speech Engine", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("Engine:    %s", engineText)),
		widget.NewLabel(fmt.Sprintf("Provider:  %s", orDefault(cfg.Provider, "local"))),
		widget.NewLabel(fmt.Sprintf("Language:  %s", orDefault(cfg.Language, "auto-detect"))),
	)

	hotkeySection := container.NewVBox(
		widget.NewLabelWithStyle("Push-to-Talk Hotkey", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		hotkeyEntry,
		saveBtn,
		saveStatus,
	)

	helpHint := widget.NewLabelWithStyle(
		"Press hotkey once to start, again to stop and transcribe.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)
	helpHint.Importance = widget.LowImportance
	helpHint.Wrapping = fyne.TextWrapWord

	info := container.NewVBox(
		engineSection,
		spacer(8),
		widget.NewSeparator(),
		spacer(8),
		hotkeySection,
		spacer(12),
		widget.NewSeparator(),
		spacer(4),
		helpHint,
	)

	w.SetContent(container.NewPadded(info))
	w.CenterOnScreen()
	w.Show()
}

func spacer(height float32) fyne.CanvasObject {
	s := canvas.NewRectangle(colorTransparent)
	s.SetMinSize(fyne.NewSize(0, height))
	return s
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
