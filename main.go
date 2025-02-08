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
)

func main() {
  var width int
  var height int
  var protocol string
  flag.StringVar(&protocol, "p", "", "Force protocol: kitty, iterm, sixel (default: auto)")
  flag.IntVar(&width, "w", 0, "Resize width (0 for no resize)")
  flag.IntVar(&height, "h", 0, "Resize height (0 for no resize)")
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
  case "": // Auto-detect
    isKittyCapable := rasterm.IsKittyCapable()
    isItermCapable := rasterm.IsItermCapable()
    isSixelCapable, err := rasterm.IsSixelCapable()
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error checking sixel capability: %v\n", err)
      return
    }
    useKitty = isKittyCapable
    useIterm = isItermCapable
    useSixel = isSixelCapable
    _, errWez := exec.LookPath("wezterm")
    _, errKit := exec.LookPath("kitty")
    if errWez == nil && useKitty {
      useKitty = false
    }
    if errKit == nil && useIterm {
      useIterm = false
    }
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
