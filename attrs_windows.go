package main

import (
  "syscall"
  "unsafe"

  "github.com/lxn/win"
)

// works everywhere
func check_device_dims() (width, height int) {
  hWnd := win.GetForegroundWindow()
  if hWnd == 0 {
    return 0, 0
  }

  // get client size
  var rect win.RECT
  if !win.GetClientRect(hWnd, &rect) {
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

var (
  kernel32 = syscall.NewLazyDLL("kernel32.dll")

  getConsoleMode = kernel32.NewProc("GetConsoleMode")
  setConsoleMode = kernel32.NewProc("SetConsoleMode")
  getStdHandle   = kernel32.NewProc("GetStdHandle")
)

const (
  STD_INPUT_HANDLE = -10

  ENABLE_ECHO_INPUT             = 0x4
  ENABLE_LINE_INPUT             = 0x2
  ENABLE_PROCESSED_INPUT        = 0x1
  ENABLE_WINDOW_INPUT           = 0x8
  ENABLE_MOUSE_INPUT            = 0x10
  ENABLE_VIRTUAL_TERMINAL_INPUT = 0x200
)

func make_raw(fd int) func() {
  // Convert fd to Windows handle
  handle := syscall.Handle(fd)

  // Get the current console mode
  var originalMode uint32
  getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&originalMode)))

  // Set raw mode by disabling processing flags
  rawMode := originalMode &^ (ENABLE_ECHO_INPUT |
    ENABLE_LINE_INPUT |
    ENABLE_MOUSE_INPUT |
    ENABLE_WINDOW_INPUT |
    ENABLE_PROCESSED_INPUT)

  // Enable virtual terminal input
  rawMode |= ENABLE_VIRTUAL_TERMINAL_INPUT

  // Apply the new mode
  setConsoleMode.Call(uintptr(handle), uintptr(rawMode))

  // Return cleanup function
  return func() {
    setConsoleMode.Call(uintptr(handle), uintptr(originalMode))
  }
}
