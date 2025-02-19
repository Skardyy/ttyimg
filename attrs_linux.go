package main

import (
  "fmt"
  "os"

  "golang.org/x/sys/unix"
)

func check_device_dims() (width, height int) {

  println("im linux")
  ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
  fmt.Println(ws)
  if err != nil {
    return 0, 0
  }
  return int(ws.Xpixel), int(ws.Ypixel)
}
