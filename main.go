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
  "github.com/boltdb/bolt"
)

var db, _ = bolt.Open("ttyimg_cache.db", 0600, nil)
var bucket_name = []byte("documents")

func main() {
  defer db.Close()
  db.Update(func(tx *bolt.Tx) error {
    tx.CreateBucket(bucket_name)
    return nil
  })
  var widthPre string
  var heightPre string
  var protocol string
  var fallback string
  var resizeMode string
  var screenSize string
  var center bool
  var cache bool
  flag.StringVar(&widthPre, "w", "80%", "Resize width: <number> (pixels) / <number>px / <number>c (cells) / <number>%")
  flag.StringVar(&heightPre, "h", "60%", "Resize height: 100 (pixels) / 100px / 100c (cells) / 100%")
  flag.StringVar(&resizeMode, "m", "Fit", "the resize mode to use when resizing: Fit, Strech, Crop")
  flag.BoolVar(&center, "center", true, "rather or not to center align the image")
  flag.StringVar(&protocol, "p", "auto", "Force protocol: kitty, iterm, sixel")
  flag.StringVar(&fallback, "f", "sixel", "fallback to when no protocol is supported: kitty, iterm, sixel")
  flag.StringVar(&screenSize, "screen", "1920x1080", "<width>x<height> or <width>x<height>xForce. specify the size of the winodw for fallback / overwrite")
  flag.BoolVar(&cache, "cache", true, "rather or not to cache the heavy operations")

  flag.Usage = func() {
    blue := "\033[34m"
    reset := "\033[0m"
    green := "\033[32m"
    purple := "\033[35m"
    yellow := "\033[33m"
    fmt.Fprintln(os.Stderr, purple+"Usage: ttyimg [options] <path_to_image>"+reset)
    order := []string{"w", "h", "m", "center", "p", "f", "screen", "cache"}
    for _, key := range order {
      f := flag.Lookup(key)
      fmt.Fprintln(os.Stderr, green+"  -"+key+reset, blue+determineType(f.DefValue)+reset)
      fmt.Fprintln(os.Stderr, "        ", flag.Lookup(key).Usage, yellow+"(default:", f.DefValue+")"+reset)
    }
  }
  flag.Parse()

  if len(flag.Args()) < 1 {
    flag.Usage()
    return
  }
  width, errWidth := ParseDimension(widthPre)
  width.direction = X
  height, errHeight := ParseDimension(heightPre)
  height.direction = Y
  if errWidth != nil || errHeight != nil {
    return
  }
  imgPath := flag.Args()[0]

  sSize := ScreenSize{}
  sSize.query(screenSize)
  resizedImg := get_img(imgPath, width, height, resizeMode, cache, sSize)

  if resizedImg == nil {
    return
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

  var offsetX int
  if center {
    offsetX, _ = CenterImage(resizedImg, sSize)
    writer.WriteString(strings.Repeat(" ", offsetX))
  }

  if useIterm {
    err := rasterm.ItermWriteImage(writer, resizedImg)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error encoding to iTerm format: %v\n", err)
      return
    }
  } else if useKitty {
    opts := rasterm.KittyImgOpts{}
    err := rasterm.KittyWriteImage(writer, resizedImg, opts)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error encoding to Kitty format: %v\n", err)
      return
    }
  } else if useSixel {
    pimg := convertToPaletted(resizedImg)
    err := rasterm.SixelWriteImage(writer, pimg)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error encoding to Sixel format: %v\n", err)
      return
    }
  } else {
    fmt.Fprintln(os.Stderr, "No capable terminal detected (Kitty, iTerm, or Sixel), and no protocol forced.")
    return
  }
  writer.WriteString("\n")
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

func detect_cap(fallback string) (iterm bool, kitty bool, sixel bool) {
  _, errWez := exec.LookPath("wezterm")
  _, errKit := exec.LookPath("kitty")

  isKittyCapable := rasterm.IsKittyCapable()
  if isKittyCapable && errKit == nil {
    return false, true, false
  }
  isItermCapable := rasterm.IsItermCapable()
  if isItermCapable && errWez == nil {
    return true, false, false
  }
  isSixelCapable := false

  switch strings.ToLower(fallback) {
  case "kitty":
    isKittyCapable = true
  case "iterm":
    isItermCapable = true
  case "sixel":
    isSixelCapable = true
  }

  return isItermCapable, isKittyCapable, isSixelCapable
}

func determineType(value string) string {
  valueLower := strings.ToLower(value)

  if valueLower == "true" || valueLower == "false" {
    return "bool"
  }

  return "string"
}
