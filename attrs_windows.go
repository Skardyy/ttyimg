package main

import (
  "fmt"

  "github.com/lxn/win"
)

func check_device_dims() (width, height int) {
  hWnd := win.GetForegroundWindow()
  if hWnd == 0 {
    fmt.Println("No foreground window found.")
    return 0, 0
  }

  var rect win.RECT
  if !win.GetWindowRect(hWnd, &rect) {
    fmt.Println("Error retrieving window rectangle.")
    return 0, 0
  }

  width = int(rect.Right - rect.Left)
  height = int(rect.Bottom - rect.Top)
  return width, height
}
