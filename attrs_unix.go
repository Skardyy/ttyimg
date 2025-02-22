//go:build linux || darwin

package main

import (
  "os"

  "golang.org/x/sys/unix"
  "golang.org/x/term"
)

// only reliable thing for linux at the moment
func getIoCtlSize() (width, height int) {
  ws, err := unix.IoctlGetWinsize(int(os.Stderr.Fd()), unix.TIOCGWINSZ)
  if err != nil {
    return 0, 0
  }

  return int(ws.Xpixel), int(ws.Ypixel)
}

func check_device_dims() (width, height int) {
  width, height = getIoCtlSize()

  return width, height
}

func make_raw(fd int) func() {
  oldstate, _ := term.MakeRaw(fd)
  return func() {
    term.Restore(fd, oldstate)
  }
}
