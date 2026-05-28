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
	"github.com/victorlui/wispr-vibe/internal/stt"
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
	w.Resize(fyne.NewSize(360, 320))
	w.CenterOnScreen()

	hotkeyCombo := cfg.Hotkey
	if hotkeyCombo == "" {
		hotkeyCombo = "Ctrl+Shift+R"
	}

	statusDot := canvas.NewCircle(colorReady)
	statusDotContainer := container.New(&fixedSizeLayout{size: fyne.NewSize(12, 12)}, statusDot)

	statusLabel := widget.NewLabelWithStyle(
		"Ready",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord
	resultLabel.Importance = widget.LowImportance

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
					statusDot.FillColor = colorError
					statusDot.Refresh()
					return
				}
				recording = true
				recordStart = time.Now()
				recordBtn.SetText("Stop")
				statusLabel.SetText("Recording")
				statusDot.FillColor = colorRecording
				statusDot.Refresh()
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
								m := int(elapsed.Minutes())
								s := int(elapsed.Seconds()) % 60
								timerLabel.SetText(fmt.Sprintf("%d:%02d", m, s))
							})
						}
					}
				}()
			} else {
				recording = false
				close(timerDone)
				recordBtn.SetText("Transcribing...")
				recordBtn.Disable()
				statusLabel.SetText("Processing")
				statusDot.FillColor = colorProcessing
				statusDot.Refresh()
				toastOverlay.SetProcessing("Transcribing...")

				go func() {
					fyne.Do(func() { w.Hide() })
					time.Sleep(300 * time.Millisecond)

					text, err := application.StopAndProcess(ctx)
					toastOverlay.Hide()

					fyne.Do(func() {
						if err != nil {
							statusLabel.SetText(fmt.Sprintf("Error: %v", err))
							statusDot.FillColor = colorError
						} else {
							statusLabel.SetText("Text inserted at cursor")
							statusDot.FillColor = colorSuccess
							resultLabel.SetText(text)
						}
						statusDot.Refresh()
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

	header := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		minimizeBtn,
		settingsBtn,
	)

	statusRow := container.NewHBox(
		statusDotContainer,
		statusLabel,
	)

	hotkeyBadge := widget.NewLabelWithStyle(
		hotkeyCombo,
		fyne.TextAlignCenter,
		fyne.TextStyle{Monospace: true},
	)
	hotkeyBadge.Importance = widget.LowImportance

	centerSection := container.NewVBox(
		container.NewCenter(timerLabel),
		spacer(6),
		container.NewCenter(recordBtn),
		spacer(4),
		container.NewCenter(hotkeyBadge),
	)

	resultPanel := container.NewVBox(
		resultLabel,
	)

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		spacer(10),
		container.NewPadded(statusRow),
		spacer(4),
		layout.NewSpacer(),
		centerSection,
		layout.NewSpacer(),
		spacer(4),
		widget.NewSeparator(),
		resultPanel,
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
	w.Resize(fyne.NewSize(380, 400))

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

	gpuCheck := widget.NewCheck("Use GPU acceleration (CUDA)", func(checked bool) {
		cfg.UseGPU = checked
		config.Save(cfg, "")
	})
	gpuCheck.SetChecked(cfg.UseGPU)

	gpuAvailable := stt.CheckGPUSupport(cfg.WhisperExePath)
	var gpuStatusText string
	var gpuStatusImportance widget.Importance
	if gpuAvailable {
		gpuStatusText = "whisper-cli has CUDA support — GPU acceleration is available."
		gpuStatusImportance = widget.SuccessImportance
	} else {
		gpuStatusText = "whisper-cli was NOT compiled with CUDA. GPU flag has no effect.\nRun build.ps1 to compile a CUDA-enabled binary."
		gpuStatusImportance = widget.WarningImportance
		gpuCheck.Disable()
	}

	gpuStatus := widget.NewLabelWithStyle(gpuStatusText, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	gpuStatus.Importance = gpuStatusImportance
	gpuStatus.Wrapping = fyne.TextWrapWord

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

	engineForm := widget.NewForm(
		widget.NewFormItem("Engine", widget.NewLabel(engineText)),
		widget.NewFormItem("Provider", widget.NewLabel(orDefault(cfg.Provider, "local"))),
		widget.NewFormItem("Language", widget.NewLabel(orDefault(cfg.Language, "auto-detect"))),
	)

	hotkeySection := container.NewVBox(
		widget.NewLabelWithStyle("Hotkey", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spacer(4),
		hotkeyEntry,
		spacer(4),
		saveBtn,
		saveStatus,
	)

	gpuSection := container.NewVBox(
		widget.NewLabelWithStyle("Performance", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spacer(4),
		gpuCheck,
		gpuStatus,
	)

	helpHint := widget.NewLabelWithStyle(
		"Press once to start recording, again to transcribe.",
		fyne.TextAlignLeading,
		fyne.TextStyle{Italic: true},
	)
	helpHint.Importance = widget.LowImportance
	helpHint.Wrapping = fyne.TextWrapWord

	info := container.NewVBox(
		widget.NewLabelWithStyle("Engine", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		spacer(2),
		engineForm,
		spacer(6),
		widget.NewSeparator(),
		spacer(8),
		hotkeySection,
		spacer(8),
		widget.NewSeparator(),
		spacer(8),
		gpuSection,
		spacer(8),
		widget.NewSeparator(),
		spacer(6),
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

type fixedSizeLayout struct {
	size fyne.Size
}

func (f *fixedSizeLayout) MinSize(_ []fyne.CanvasObject) fyne.Size {
	return f.size
}

func (f *fixedSizeLayout) Layout(objects []fyne.CanvasObject, _ fyne.Size) {
	for _, o := range objects {
		o.Resize(f.size)
		o.Move(fyne.NewPos(0, 0))
	}
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
