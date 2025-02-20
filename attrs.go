package main

import (
  "bufio"
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

func get_size_cells(cellHandler *string) (int, int, error) {
  response, err := queryTerminal("\033[18t")
  if err != nil {
    fd := int(os.Stdout.Fd())
    widthCell, heightCell, err := term.GetSize(fd)
    *cellHandler = "go term"
    return widthCell, heightCell, err
  }

  //[8;36;172t
  parts := strings.Split(response, ";")
  height, _ := strconv.Atoi(parts[1])
  width, _ := strconv.Atoi(strings.Replace(parts[2], "t", "", 1))
  *cellHandler = "osc"

  return width, height, nil
}

func (s *ScreenSize) query(fallbackPx string, fallbackCell string) {
  forcePx := strings.Contains(strings.ToLower(fallbackPx), "force")
  forceCell := strings.Contains(strings.ToLower(fallbackCell), "force")
  hanlderPx := ""
  handlerCell := ""

  // attempt to query px when not forced
  if !forcePx {
    var err error
    s.widthPx, s.heightPx, err = get_size_osc()
    hanlderPx = "osc"
    if err != nil {
      s.widthPx, s.heightPx = check_device_dims()
      hanlderPx = "win api or ioctl"
    }
  }

  // forced or failed to query px
  if s.widthPx == 0 || forcePx {
    parts := strings.Split(fallbackPx, "x")
    s.widthPx, _ = strconv.Atoi(parts[0])
    s.heightPx, _ = strconv.Atoi(parts[1])
    hanlderPx = "fallback"
  }

  var err error
  // forced or failed to query cells
  s.widthCell, s.heightCell, err = get_size_cells(&handlerCell)
  if err != nil || forceCell {
    parts := strings.Split(fallbackCell, "x")
    s.widthCell, _ = strconv.Atoi(parts[0])
    s.heightCell, _ = strconv.Atoi(parts[1])
    handlerCell = "fallback"
  }

  logMsg := fmt.Sprintf("px handler: <%s> gave %dx%d\n    cell handler: <%s> gave %dx%d\n    forcePx: %t\n    forceCell: %t", hanlderPx, s.widthPx, s.heightPx, handlerCell, s.widthCell, s.heightCell, forcePx, forceCell)
  logger.Write(logMsg)
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
  clean_func := make_raw(fd)
  defer clean_func()

  fmt.Fprintf(os.Stdout, escapeSeq)

  ch := make(chan string, 1)
  go func() {
    reader := bufio.NewReader(os.Stdin)
    response, _ := reader.ReadString('t')
    ch <- response
  }()

  select {
  case response := <-ch:
    return response, nil
  case <-time.After(50 * time.Millisecond):
    return "", fmt.Errorf("timeout waiting for terminal response")
  }
}
