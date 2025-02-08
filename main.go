package main

import (
  "bufio"
  "flag"
  "fmt"
  "image"
  "image/color/palette"
  "image/draw"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "os/exec"
  "strings"

  "github.com/BourgeoisBear/rasterm"
  "github.com/nfnt/resize"
  "golang.org/x/term"
)

func main() {
  var width int
  var height int
  var protocol string
  var fallback string
  flag.StringVar(&protocol, "p", "auto", "Force protocol: kitty, iterm, sixel")
  flag.StringVar(&fallback, "f", "none", "fallback to when no protocol is supported: kitty, iterm, sixel")
  flag.IntVar(&width, "w", 0, "Resize width")
  flag.IntVar(&height, "h", 0, "Resize height")
  flag.Parse()

  if len(flag.Args()) < 1 {
    fmt.Fprintln(os.Stderr, "Usage: img_encoder [options] <path_to_image>")
    flag.PrintDefaults()
    return
  }
  imgPath := flag.Args()[0]

  imgFile, err := os.Open(imgPath)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening image: %v\n", err)
    return
  }
  defer imgFile.Close()

  img, _, err := image.Decode(imgFile)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error decoding image: %v\n", err)
    return
  }

  originalWidth := img.Bounds().Dx()
  originalHeight := img.Bounds().Dy()

  resizedWidth := width
  resizedHeight := height

  if resizedWidth > 0 && resizedHeight == 0 {
    resizedHeight = int(float64(originalHeight) * (float64(resizedWidth) / float64(originalWidth)))
  }

  resizedImg := img
  if resizedWidth > 0 || resizedHeight > 0 {
    if resizedWidth == 0 {
      resizedWidth = originalWidth
    }
    if resizedHeight == 0 {
      resizedHeight = originalHeight
    }
    resizedImg = resize.Resize(uint(resizedWidth), uint(resizedHeight), img, resize.Lanczos3)
  }

  useKitty := false
  useIterm := false
  useSixel := false

  switch strings.ToLower(protocol) {
  case "kitty":
    useKitty = true
  case "iterm":
    useIterm = true
  case "sixel":
    useSixel = true
  case "auto": // Auto-detect
    useIterm, useKitty, useSixel = detect_cap(fallback)
  default:
    fmt.Fprintf(os.Stderr, "Error: invalid protocol '%s'. Must be kitty, iterm, or sixel.\n", protocol)
    flag.PrintDefaults()
    return
  }

  writer := NewBufferedWriter()
  defer writer.Flush()

  if useIterm {
    err = rasterm.ItermWriteImage(writer, resizedImg)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error encoding to iTerm format: %v\n", err)
      return
    }
  } else if useKitty {
    opts := rasterm.KittyImgOpts{}
    err = rasterm.KittyWriteImage(writer, resizedImg, opts)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error encoding to Kitty format: %v\n", err)
      return
    }
  } else if useSixel {
    pimg := convertToPaletted(resizedImg)
    err = rasterm.SixelWriteImage(writer, pimg)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error encoding to Sixel format: %v\n", err)
      return
    }
  } else {
    fmt.Fprintln(os.Stderr, "No capable terminal detected (Kitty, iTerm, or Sixel), and no protocol forced.")
    return
  }
}

func convertToPaletted(img image.Image) *image.Paletted {
  bounds := img.Bounds()

  paletted := image.NewPaletted(bounds, palette.WebSafe)
  draw.FloydSteinberg.Draw(paletted, bounds, img, bounds.Min)

  return paletted
}

func NewBufferedWriter() *bufio.Writer {
  return bufio.NewWriterSize(os.Stdout, 64*1024) // 64 KB buffer
}

func checkDeviceAttrs() (bool, error) {
  oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
  if err != nil {
    return false, err
  }
  defer term.Restore(int(os.Stdin.Fd()), oldState)

  // Send device attributes query
  fmt.Fprint(os.Stdout, "\x1b[c")
  os.Stdout.Sync()

  response := ""
  buf := make([]byte, 1)

  // Read response until we get 'c'
  for {
    _, err := os.Stdin.Read(buf)
    if err != nil {
      return false, err
    }
    response += string(buf)
    if buf[0] == 'c' {
      break
    }
  }

  return strings.Contains(response, ";4;") || strings.Contains(response, ";4c"), nil
}

func detect_cap(fallback string) (iterm bool, kitty bool, sixel bool) {
  _, errWez := exec.LookPath("wezterm imgcat")
  if errWez == nil {
    return true, false, false
  }

  _, errKit := exec.LookPath("kitty icat")
  if errKit == nil {
    return false, true, false
  }

  isKittyCapable := rasterm.IsKittyCapable()
  isItermCapable := rasterm.IsItermCapable()
  isSixelCapable, _ := rasterm.IsSixelCapable()

  if !isKittyCapable && !isItermCapable && !isSixelCapable {
    if flag, _ := checkDeviceAttrs(); flag {
      isSixelCapable = true
    } else {
      switch strings.ToLower(fallback) {
      case "kitty":
        isKittyCapable = true
      case "iterm":
        isItermCapable = true
      case "sixel":
        isSixelCapable = true
      }
    }
  }

  return isItermCapable, isKittyCapable, isSixelCapable
}
