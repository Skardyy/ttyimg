package main

import (
  "fmt"

  "github.com/lxn/win"
)

// works everywhere
func check_device_dims() (width, height int) {
  hWnd := win.GetForegroundWindow()
  if hWnd == 0 {
    fmt.Println("No foreground window found.")
    return 0, 0
  }

  // get client size
  var rect win.RECT
  if !win.GetClientRect(hWnd, &rect) {
    fmt.Println("Error retrieving window rectangle.")
    return 0, 0
  }

  // get frame size
  style := win.GetWindowLong(win.GetForegroundWindow(), win.GWL_STYLE)
  frameRect := win.RECT{
    Left:   0,
    Right:  0,
    Bottom: 0,
    Top:    0,
  }
  win.AdjustWindowRect(&frameRect, uint32(style), false)
  frame_height := frameRect.Bottom - frameRect.Top
  frame_width := frameRect.Right - frameRect.Left

  logicalWidth := int(rect.Right-rect.Left) - int(frame_width)
  logicalHeight := int(rect.Bottom-rect.Top) - int(frame_height)

  return logicalWidth, logicalHeight
}
