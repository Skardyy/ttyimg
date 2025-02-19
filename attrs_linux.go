package main

import (
  "fmt"
  "os"
  "os/exec"
  "strconv"
  "strings"

  "github.com/gotk3/gotk3/gtk"
  "github.com/therecipe/qt/widgets"
)

// Method 1: Using xwininfo (X11)
func getTerminalSizeX11() (width, height int, err error) {
  // Get window ID of current terminal
  cmd := exec.Command("xdotool", "getactivewindow")
  output, err := cmd.Output()
  if err != nil {
    return 0, 0, fmt.Errorf("failed to get window ID: %v", err)
  }
  windowID := strings.TrimSpace(string(output))

  // Get window geometry
  cmd = exec.Command("xwininfo", "-id", windowID)
  output, err = cmd.Output()
  if err != nil {
    return 0, 0, fmt.Errorf("failed to get window info: %v", err)
  }

  // Parse output
  for _, line := range strings.Split(string(output), "\n") {
    if strings.Contains(line, "Width:") {
      width, _ = strconv.Atoi(strings.Fields(line)[1])
    }
    if strings.Contains(line, "Height:") {
      height, _ = strconv.Atoi(strings.Fields(line)[1])
    }
  }

  return width, height, nil
}

// Method 2: Using escape sequences
func getTerminalSizeEscapeSeq() (width, height int, err error) {
  // Save cursor position
  fmt.Print("\x1b[s")
  // Move cursor to bottom-right
  fmt.Print("\x1b[999C\x1b[999B")
  // Get cursor position
  fmt.Print("\x1b[6n")
  // Restore cursor position
  fmt.Print("\x1b[u")

  var buf [64]byte
  n, err := os.Stdin.Read(buf[:])
  if err != nil {
    return 0, 0, err
  }

  // Parse response of the form "\x1b[rows;colsR"
  response := string(buf[:n])
  if !strings.HasPrefix(response, "\x1b[") || !strings.HasSuffix(response, "R") {
    return 0, 0, fmt.Errorf("invalid response")
  }

  parts := strings.Split(response[2:len(response)-1], ";")
  if len(parts) != 2 {
    return 0, 0, fmt.Errorf("invalid response format")
  }

  height, _ = strconv.Atoi(parts[0])
  width, _ = strconv.Atoi(parts[1])

  return width, height, nil
}

// Method 3: Using GTK
func getTerminalSizeGTK() (width, height int, err error) {
  gtk.Init(nil)
  defer gtk.Main()

  // Get the default display
  display, err := gtk.DisplayGetDefault()
  if err != nil {
    return 0, 0, fmt.Errorf("failed to get display: %v", err)
  }

  // Get active window
  screen, err := display.GetDefaultScreen()
  if err != nil {
    return 0, 0, fmt.Errorf("failed to get screen: %v", err)
  }

  window := screen.GetActiveWindow()
  if window == nil {
    return 0, 0, fmt.Errorf("no active window found")
  }

  width, height = window.GetSize()
  return width, height, nil
}

// Method 4: Using Qt
func getTerminalSizeQt() (width, height int, err error) {
  // Initialize Qt application
  app := widgets.NewQApplication(len(os.Args), os.Args)
  defer app.Quit()

  // Get active window
  activeWindow := app.ActiveWindow()
  if activeWindow == nil {
    return 0, 0, fmt.Errorf("no active window found")
  }

  // Get size
  width = activeWindow.Width()
  height = activeWindow.Height()

  return width, height, nil
}

// Method 5: Using wmctrl (another X11 tool)
func getTerminalSizeWmctrl() (width, height int, err error) {
  cmd := exec.Command("wmctrl", "-l", "-G")
  output, err := cmd.Output()
  if err != nil {
    return 0, 0, fmt.Errorf("failed to execute wmctrl: %v", err)
  }

  // Get current window ID
  currentWindow, err := exec.Command("xdotool", "getactivewindow").Output()
  if err != nil {
    return 0, 0, fmt.Errorf("failed to get active window: %v", err)
  }
  windowID := strings.TrimSpace(string(currentWindow))

  // Parse wmctrl output to find our window
  for _, line := range strings.Split(string(output), "\n") {
    fields := strings.Fields(line)
    if len(fields) >= 7 && strings.Contains(fields[0], windowID) {
      width, _ = strconv.Atoi(fields[4])
      height, _ = strconv.Atoi(fields[5])
      return width, height, nil
    }
  }

  return 0, 0, fmt.Errorf("window not found in wmctrl output")
}

func check_device_dims() (width, height int) {
  // Try all methods in sequence until one works
  methods := []struct {
    name string
    fn   func() (int, int, error)
  }{
    {"X11", getTerminalSizeX11},
    {"GTK", getTerminalSizeGTK},
    {"Qt", getTerminalSizeQt},
    {"wmctrl", getTerminalSizeWmctrl},
    {"Escape Sequence", getTerminalSizeEscapeSeq},
  }

  for _, method := range methods {
    if w, h, err := method.fn(); err == nil && w > 0 && h > 0 {
      fmt.Printf("Successfully got terminal size using %s method\n", method.name)
      return w, h
    }
  }

  fmt.Println("All methods failed, returning default values")
  return 0, 0
}
