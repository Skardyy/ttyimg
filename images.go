package main

import (
  "bytes"
  "fmt"
  "image"
  "image/color"
  "image/draw"
  "math"
  "os"
  "strings"

  "github.com/nfnt/resize"
  "github.com/srwiley/oksvg"
  "github.com/srwiley/rasterx"
  "golang.org/x/image/tiff"
  "golang.org/x/image/webp"
)

// doesn't work currently need to figure out read that
func getTerminalSize() (width int, height int, err error) {
  os.Stdout.Write([]byte("\x1b[14t"))
  return 0, 0, nil
}

func get_resize_mode(resizeMode string) ResizeMethod {
  if strings.ToLower(resizeMode) == "fit" {
    return Fit
  }
  if strings.ToLower(resizeMode) == "strech" {
    return Stretch
  }
  if strings.ToLower(resizeMode) == "crop" {
    return Crop
  }
  return Fit
}
func get_img(path string, width int, height int, resizeMod string) image.Image {
  var img image.Image

  if strings.Contains(path, ".pdf") {
    width
  } else {

    imgFile, err := os.Open(path)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error opening image: %v\n", err)
      return nil
    }
    defer imgFile.Close()
    img = get_content(imgFile, width, height)
  }

  resizeMode := get_resize_mode(resizeMod)
  resizedImg, _ := ResizeImage(img, uint(width), uint(height), resizeMode)
  return resizedImg
}

func get_content(file *os.File, width int, height int) image.Image {
  name := file.Name()
  if height == width && width == 0 {
    width = 200
    height = 200
  } else {
    bigger := width
    if height > bigger {
      bigger = height
    }
    width = bigger
    height = bigger
  }

  if strings.Contains(name, ".svg") {
    buf := new(bytes.Buffer)
    _, err := buf.ReadFrom(file)
    if err != nil {
      return nil
    }

    icon, err := oksvg.ReadIconStream(buf)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error reading svg: %v\n", err)
      return nil
    }
    icon.SetTarget(0, 0, float64(width), float64(height))
    img := image.NewRGBA(image.Rect(0, 0, width, height))
    draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
    raster := rasterx.NewDasher(width, height, rasterx.NewScannerGV(width, height, img, img.Bounds()))
    icon.Draw(raster, 1.0)

    return img
  } else if strings.Contains(name, ".tif") {
    img, err := tiff.Decode(file)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error decoding tiff: %v\n", err)
      return nil
    }
    return img
  } else if strings.Contains(name, ".webp") {
    img, err := webp.Decode(file)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error decoding webp: %v\n", err)
      return nil
    }
    return img
  } else {
    img, _, err := image.Decode(file)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error decoding image: %v\n", err)
      return nil
    }
    return img
  }
}

type ResizeMethod string

const (
  Stretch ResizeMethod = "Stretch"
  Crop    ResizeMethod = "Crop"
  Fit     ResizeMethod = "Fit"
)

func computeDimensions(origW, origH int, width, height uint) (uint, uint) {
  if width == 0 && height == 0 {
    return uint(origW), uint(origH)
  }
  if width != 0 && height != 0 {
    return uint(width), uint(height)
  }
  if width == 0 {
    return uint(math.Round(float64(height) * (float64(origW) / float64(origH)))), height
  }
  if height == 0 {
    return width, uint(math.Round(float64(width) * (float64(origH) / float64(origW))))
  }
  return width, height
}

func ResizeImage(img image.Image, width, height uint, method ResizeMethod) (image.Image, error) {
  bounds := img.Bounds()
  srcWidth := bounds.Dx()
  srcHeight := bounds.Dy()
  width, height = computeDimensions(srcWidth, srcHeight, width, height)
  switch method {
  case Fit:
    return resize.Thumbnail(width, height, img, resize.Lanczos3), nil
  case Stretch:
    // Resize without preserving the aspect ratio
    return resize.Resize(width, height, img, resize.Lanczos3), nil
  case Crop:
    // Crop the image to the specified dimensions
    cropped := cropImage(img, int(width), int(height))
    return cropped, nil
  default:
    return nil, fmt.Errorf("unsupported resize method: %s", method)
  }
}

// cropImage crops the image to the specified width and height
func cropImage(img image.Image, width, height int) image.Image {
  srcBounds := img.Bounds()
  srcWidth := srcBounds.Dx()
  srcHeight := srcBounds.Dy()

  // Calculate cropping dimensions
  var cropWidth, cropHeight int
  if float64(srcWidth)/float64(width) > float64(srcHeight)/float64(height) {
    cropWidth = int(float64(srcHeight) * float64(width) / float64(height))
    cropHeight = srcHeight
  } else {
    cropWidth = srcWidth
    cropHeight = int(float64(srcWidth) * float64(height) / float64(width))
  }

  // Calculate cropping offsets
  offsetX := (srcWidth - cropWidth) / 2
  offsetY := (srcHeight - cropHeight) / 2

  // Crop the image
  cropped := img.(interface {
    SubImage(r image.Rectangle) image.Image
  }).SubImage(image.Rect(offsetX, offsetY, offsetX+cropWidth, offsetY+cropHeight))

  // Resize the cropped image to the target dimensions
  return resize.Resize(uint(width), uint(height), cropped, resize.Lanczos3)
}
