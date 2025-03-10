package main

import (
  "bytes"
  "fmt"
  "image"
  "image/color"
  "image/draw"
  "image/gif"
  "io"
  "math"
  "os"
  "os/exec"
  "path/filepath"
  "strings"

  "github.com/boltdb/bolt"
  "github.com/srwiley/oksvg"
  "github.com/srwiley/rasterx"
  "golang.org/x/image/bmp"
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

func command_exists(cmd string) bool {
  _, err := exec.LookPath(cmd)
  return err == nil
}

func libre_command(path string, tmpDir string) (*exec.Cmd, bool) {
  if command_exists("libreoffice") {
    cmd := exec.Command(
      "libreoffice",
      "--headless",
      "--convert-to",
      "png",
      path,
      "--outdir",
      tmpDir,
    )
    return cmd, true
  }
  if command_exists("soffice") {
    cmd := exec.Command(
      "soffice",
      "--headless",
      "--convert-to",
      "png",
      path,
      "--outdir",
      tmpDir,
    )
    return cmd, true
  }

  return nil, false
}

func is_special_doc(path string, width int, height int, should_cache bool) (image.Image, bool) {
  key := []byte(path)
  var cachedImage []byte
  // get cache
  if should_cache {
    db.View(func(tx *bolt.Tx) error {
      cachedImage = tx.Bucket(bucket_name).Get(key)
      return nil
    })
    if cachedImage != nil {
      return bytesToImage(cachedImage), true
    }
  }

  exts := []string{".pdf", ".xls", ".doc", ".ppt", ".ods", ".odp", ".odg", ".odt"}
  for _, ext := range exts {
    if strings.Contains(path, ext) {
      tmpDir, _ := os.MkdirTemp("", "tmp")
      cmd, libre_exists := libre_command(path, tmpDir)
      if libre_exists {
        cmd.Run()
      } else {
        return nil, false
      }

      tmpFile := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) + ".png"
      new_path := filepath.Join(tmpDir, tmpFile)
      img := read_img(new_path, width, height)
      // put into cache
      if should_cache {
        db.Update(func(tx *bolt.Tx) error {
          tx.Bucket(bucket_name).Put(key, imageToBytes(img))
          return nil
        })
      }
      return img, true
    }
  }

  return nil, true
}

func read_img(path string, width, height int) image.Image {
  imgFile, err := os.Open(path)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening image: %v\n", err)
    return nil
  }
  defer imgFile.Close()
  return get_content(imgFile, width, height)
}

func get_img(path string, widthDm Dimension, heightDm Dimension, resizeMod string, cache bool, sSize ScreenSize) image.Image {
  var img image.Image
  width, height := widthDm.GetPixel(sSize), heightDm.GetPixel(sSize)

  img, backend_exists := is_special_doc(path, width, height, cache)
  if !backend_exists {
    fmt.Fprintln(os.Stderr, "can't preview documents, no supported backend is installed")
    return nil
  } else if img == nil {
    img = read_img(path, width, height)
  }

  resizeMode := get_resize_mode(resizeMod)
  resizedImg, _ := ResizeImage(img, uint(width), uint(height), resizeMode)
  return resizedImg
}

func decodeSVG(file *os.File, width, height int) (image.Image, error) {
  buf := new(bytes.Buffer)
  _, err := buf.ReadFrom(file)
  if err != nil {
    return nil, fmt.Errorf("Error reading SVG file: %v", err)
  }

  icon, err := oksvg.ReadIconStream(buf)
  if err != nil {
    return nil, fmt.Errorf("Error reading SVG icon: %v", err)
  }

  // Set the target size for the SVG rendering
  icon.SetTarget(0, 0, float64(width), float64(height))

  // Create a new image and draw the SVG content onto it
  img := image.NewRGBA(image.Rect(0, 0, width, height))
  draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

  raster := rasterx.NewDasher(width, height, rasterx.NewScannerGV(width, height, img, img.Bounds()))
  icon.Draw(raster, 1.0)

  return img, nil
}

func decodeImage(file *os.File, width, height int) (image.Image, error) {
  name := file.Name()
  formatDecoders := map[string]func(io.Reader) (image.Image, error){
    ".tif":  tiff.Decode,
    ".webp": webp.Decode,
    ".bmp":  bmp.Decode,
    ".gif":  gif.Decode,
  }

  for ext, decodeFunc := range formatDecoders {
    if strings.Contains(name, ext) {
      img, err := decodeFunc(file)
      if err != nil {
        return nil, fmt.Errorf("Error decoding %s: %v", ext, err)
      }
      return img, nil
    }
  }

  // Handle SVG files
  if strings.Contains(name, ".svg") {
    img, err := decodeSVG(file, width, height)
    if err != nil {
      return nil, fmt.Errorf("Error decoding SVG: %v", err)
    }
    return img, nil
  }

  img, _, err := image.Decode(file)
  if err != nil {
    return nil, fmt.Errorf("Error decoding image with no matching extension: %v", err)
  }
  return img, nil
}

func get_content(file *os.File, width, height int) image.Image {
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

  img, err := decodeImage(file, width, height)
  if err != nil {
    fmt.Fprintln(os.Stderr, "Error decoding image")
    return nil
  }
  return img
}

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
