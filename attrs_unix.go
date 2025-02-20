//go:build linux || darwin

package main

import (
  "golang.org/x/sys/unix"
  "os"
)

// only reliable thing for linux at the moment
func getIoCtlSize() (width, height int) {
  ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
  if err != nil {
    return 0, 0
  }

  return int(ws.Xpixel), int(ws.Ypixel)
}

func check_device_dims() (width, height int) {
  width, height = getIoCtlSize()

  return width, height
}

//TODO query from compositors the size
// may not be viable in wayland
