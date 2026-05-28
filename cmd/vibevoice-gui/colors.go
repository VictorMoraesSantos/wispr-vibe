package main

import "image/color"

var (
	colorReady       = color.NRGBA{R: 74, G: 210, B: 135, A: 255}  // mint green
	colorRecording   = color.NRGBA{R: 240, G: 75, B: 75, A: 255}   // warm red
	colorProcessing  = color.NRGBA{R: 129, G: 110, B: 240, A: 255} // violet (matches primary)
	colorSuccess     = color.NRGBA{R: 74, G: 210, B: 135, A: 255}  // mint green
	colorError       = color.NRGBA{R: 240, G: 75, B: 75, A: 255}   // warm red
	colorTransparent = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	colorSurface     = color.NRGBA{R: 26, G: 26, B: 32, A: 255}    // raised surface
	colorDimText     = color.NRGBA{R: 140, G: 138, B: 152, A: 255} // secondary text
)
