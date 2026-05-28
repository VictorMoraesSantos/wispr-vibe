//go:build !windows

package toast

import "fmt"

type Toast struct{}

func New() *Toast                       { return &Toast{} }
func (t *Toast) Show(text string)       {}
func (t *Toast) Hide()                  {}
func (t *Toast) SetText(text string)    {}
func (t *Toast) SetProcessing(text string) {}
func (t *Toast) Destroy()               {}

func FormatRecordingText(seconds int) string {
	return fmt.Sprintf("Recording  %ds", seconds)
}
