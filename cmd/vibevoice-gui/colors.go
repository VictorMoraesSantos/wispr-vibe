package main

import "image/color"

var (
	colorReady      = color.NRGBA{R: 74, G: 222, B: 128, A: 255}  // soft green
	colorRecording  = color.NRGBA{R: 239, G: 68, B: 68, A: 255}   // vivid red
	colorProcessing = color.NRGBA{R: 99, G: 102, B: 241, A: 255}  // indigo (matches primary)
	colorSuccess    = color.NRGBA{R: 74, G: 222, B: 128, A: 255}  // soft green
	colorError      = color.NRGBA{R: 239, G: 68, B: 68, A: 255}   // vivid red
	colorTransparent = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
)
