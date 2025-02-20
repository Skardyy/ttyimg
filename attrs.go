package main

import (
  "fmt"
  "os"
  "regexp"
  "strconv"
  "strings"
  "time"

  "golang.org/x/term"
)

type DimensionType int
type DimensionDirection int

const (
  Pixel DimensionType = iota
  Cell
  Percent
)

const (
  X DimensionDirection = iota
  Y
)

type Dimension struct {
  value     int
  kind      DimensionType
  direction DimensionDirection
}

type ScreenSize struct {
  widthPx    int
  heightPx   int
  widthCell  int
  heightCell int
}

func get_size_osc() (int, int, error) {
  response, err := queryTerminal("\033[14t")
  if err != nil {
    return 0, 0, err
  }

  //\x1b[4;680;1550t
  parts := strings.Split(response, ";")
  height, _ := strconv.Atoi(parts[1])
  width, _ := strconv.Atoi(strings.Replace(parts[2], "t", "", 1))

  return width, height, nil
}

func (s *ScreenSize) query(fallback string) {
  force := strings.Contains(strings.ToLower(fallback), "force")

  // attempt to query when not forced
  if !force {
    var err error
    s.widthPx, s.heightPx, err = get_size_osc()
    if err != nil {
      s.widthPx, s.heightPx = check_device_dims()
    }
  }

  // forced or failed to query
  if s.widthPx == 0 || force {
    parts := strings.Split(fallback, "x")
    s.widthPx, _ = strconv.Atoi(parts[0])
    s.heightPx, _ = strconv.Atoi(parts[1])
  }
  fd := int(os.Stdout.Fd())
  s.widthCell, s.heightCell, _ = term.GetSize(fd)
}

func (dm *Dimension) GetPixel(screenSize ScreenSize) int {
  if dm.value == 0 {
    return 0
  }

  var sizePx, sizeCell int
  if dm.direction == X {
    sizePx, sizeCell = screenSize.widthPx, screenSize.widthCell
  } else {
    sizePx, sizeCell = screenSize.heightPx, screenSize.heightCell
  }

  switch dm.kind {
  case Pixel:
    // already pixel
    return dm.value
  case Cell:
    // screen pixel / screen cell * value
    return sizePx / sizeCell * dm.value
  case Percent:
    // screen pixel / (dm.value / 100)
    normalizedPercent := float32(dm.value) / 100
    value := float32(sizePx) * normalizedPercent
    return int(value)
  }

  return 0
}

func ParseDimension(input string) (Dimension, error) {
  var dimension Dimension
  input = strings.ToLower(input)

  numericRegex := regexp.MustCompile(`^(-?\d+)(px|c|%)?$`)
  matches := numericRegex.FindStringSubmatch(input)

  if matches == nil {
    return dimension, fmt.Errorf("invalid dimension format: %s", input)
  }

  value, err := strconv.Atoi(matches[1])
  if err != nil {
    return dimension, err
  }

  dimension.value = value
  switch matches[2] {
  case "px", "":
    dimension.kind = Pixel
  case "c":
    dimension.kind = Cell
  case "%":
    dimension.kind = Percent
  default:
    return dimension, fmt.Errorf("unknown dimension type: %s", matches[2])
  }

  return dimension, nil
}

// sends osc and waits max 200ms for the res
func queryTerminal(escapeSeq string) (string, error) {
  fd := int(os.Stdin.Fd())
  oldState, err := term.MakeRaw(fd)
  if err != nil {
    return "", err
  }
  defer term.Restore(fd, oldState)

  fmt.Fprint(os.Stdout, escapeSeq)

  responseChan := make(chan string, 1)

  go func() {
    buf := make([]byte, 32)
    n, err := os.Stdin.Read(buf)
    if err == nil {
      responseChan <- string(buf[:n])
    } else {
      responseChan <- ""
    }
  }()

  select {
  case response := <-responseChan:
    return response, nil
  case <-time.After(200 * time.Millisecond):
    return "", fmt.Errorf("timeout waiting for response")
  }
}
