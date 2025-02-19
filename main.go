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
  var width int
  var height int
  var protocol string
  var fallback string
  var resizeMode string
  var cache bool
  flag.StringVar(&protocol, "p", "auto", "Force protocol: kitty, iterm, sixel")
  flag.StringVar(&fallback, "f", "sixel", "fallback to when no protocol is supported: kitty, iterm, sixel")
  flag.StringVar(&resizeMode, "m", "Fit", "the resize mode to use when resizing: Fit, Strech, Crop")
  flag.BoolVar(&cache, "c", true, "rather or not to cache the heavy operations")
  flag.IntVar(&width, "w", 0, "Resize width")
  flag.IntVar(&height, "h", 0, "Resize height")
  flag.Parse()

  if len(flag.Args()) < 1 {
    fmt.Fprintln(os.Stderr, "Usage: ttyimg [options] <path_to_image>")
    flag.PrintDefaults()
    return
  }
  imgPath := flag.Args()[0]

  resizedImg := get_img(imgPath, width, height, resizeMode, cache)

  if resizedImg == nil {
    return
  }

  useKitty := false
  useIterm := false
  useSixel := false

  // tw, th := check_device_dims()
  // println(tw, th)

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
