package main

import (
  "fmt"
  "os"
  "strings"

  "golang.org/x/term"
)

func checkDeviceDims() {
  res, _ := read_osc_seq("\033[14t")

  //"[4;680;1550t"
  parts := strings.Split(res, ";")
  if len(parts) > 1 {

    height := parts[1]
    width := strings.Replace(parts[2], "t", "", 1)
    fmt.Println("width: ", width, " height: ", height)
  }
}

func read_osc_seq(osc string) (string, error) {
  fd := int(os.Stdin.Fd())

  oldState, err := term.MakeRaw(fd)
  if err != nil {
    return "", err
  }
  defer term.Restore(fd, oldState)
  fmt.Print(osc)

  var buf []byte
  singleByte := make([]byte, 1)
  for {
    n, err := os.Stdin.Read(singleByte)
    if err != nil {
      return "", err
    }
    if n > 0 {
      buf = append(buf, singleByte[0])
      if singleByte[0] == 't' {
        break
      }
    }
  }

  if len(buf) > 0 && buf[0] == 0x1b {
    return string(buf[1:]), nil
  }
  return string(buf), nil
}
