package main

import (
  "bytes"
  "fmt"
  "image"
  "image/png"

  "github.com/nfnt/resize"
)

type ResizeMethod string

const (
  Stretch ResizeMethod = "Stretch"
  Crop    ResizeMethod = "Crop"
  Fit     ResizeMethod = "Fit"
)

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

func ResizeImage(img image.Image, width, height uint, method ResizeMethod) (image.Image, error) {
  if img == nil {
    return nil, fmt.Errorf("error decoding image")
  }
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

func imageToBytes(img image.Image) []byte {
  buf := bytes.Buffer{}
  png.Encode(&buf, img)

  return buf.Bytes()
}
func bytesToImage(bufBytes []byte) image.Image {
  buf := bytes.NewBuffer(bufBytes)
  img, _ := png.Decode(buf)
  return img
}

func CenterImage(img image.Image, sSize ScreenSize) (offsetX, offsetY int) {
  bounds := img.Bounds()
  imgW := bounds.Dx()
  imgH := bounds.Dy()
  // in px
  offsetX = (sSize.widthPx - imgW) / 2
  offsetY = (sSize.heightPx - imgH) / 2
  // in cells
  offsetX = offsetX / (sSize.widthPx / sSize.widthCell)
  offsetY = offsetY / (sSize.heightPx / sSize.heightCell)

  if offsetY < 0 {
    offsetY = 0
  }
  if offsetX < 0 {
    offsetX = 0
  }
  return
}
