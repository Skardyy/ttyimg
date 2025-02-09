package main

import (
  "bytes"
  "fmt"
  "image"
  "image/color"
  "image/draw"
  "os"
  "strings"

  "github.com/nfnt/resize"
  "github.com/srwiley/oksvg"
  "github.com/srwiley/rasterx"
  "golang.org/x/image/tiff"
  "golang.org/x/image/webp"
)

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

  imgFile, err := os.Open(path)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening image: %v\n", err)
    return nil
  }
  defer imgFile.Close()
  img = get_content(imgFile, width, height)

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
    if width == 0 {
      width = height
    }
    if height == 0 {
      height = width
    }
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
  } else if strings.Contains(name, ".tiff") {
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

func ResizeImage(img image.Image, width, height uint, method ResizeMethod) (image.Image, error) {
  switch method {
  case Stretch:
    // Resize without preserving the aspect ratio
    return resize.Resize(width, height, img, resize.Lanczos3), nil
  case Crop:
    // Crop the image to the specified dimensions
    cropped := cropImage(img, int(width), int(height))
    return cropped, nil
  case Fit:
    bounds := img.Bounds()
    srcWidth := uint(bounds.Dx())
    srcHeight := uint(bounds.Dy())
    srcAspectRatio := float64(srcWidth) / float64(srcHeight)

    var newWidth, newHeight uint

    // Calculate both possible dimensions
    option1Width := width
    option1Height := uint(float64(width) / srcAspectRatio)

    option2Width := uint(float64(height) * srcAspectRatio)
    option2Height := height

    // Choose the option that results in the larger area
    area1 := option1Width * option1Height
    area2 := option2Width * option2Height

    if area1 >= area2 {
      newWidth = option1Width
      newHeight = option1Height
    } else {
      newWidth = option2Width
      newHeight = option2Height
    }

    // Perform the resize
    return resize.Resize(newWidth, newHeight, img, resize.Lanczos3), nil
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
