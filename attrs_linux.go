package main

import (
  "fmt"
  "golang.org/x/sys/unix"
  "os"
  "os/exec"
  "regexp"
  "strconv"
  "strings"
  "syscall"
  "time"
  "unsafe"
)

type TermSize struct {
  Width  int
  Height int
  Source string // Track which method succeeded
}

// Wayland native approach
func getWaylandSize() (*TermSize, error) {
  // Check if we're in Wayland
  if os.Getenv("WAYLAND_DISPLAY") == "" {
    return nil, fmt.Errorf("not in Wayland session")
  }

  // Try using wlr-randr for wlroots-based compositors
  if size, err := getWlrRandrSize(); err == nil {
    return size, nil
  }

  // Try using swaymsg for Sway
  if size, err := getSwaySize(); err == nil {
    return size, nil
  }

  return nil, fmt.Errorf("no Wayland method succeeded")
}

func getWlrRandrSize() (*TermSize, error) {
  cmd := exec.Command("wlr-randr")
  output, err := cmd.Output()
  if err != nil {
    return nil, fmt.Errorf("wlr-randr error: %v", err)
  }

  // Parse wlr-randr output
  re := regexp.MustCompile(`(\d+)x(\d+)`)
  match := re.FindSubmatch(output)
  if match != nil {
    width, _ := strconv.Atoi(string(match[1]))
    height, _ := strconv.Atoi(string(match[2]))
    return &TermSize{Width: width, Height: height, Source: "wlr-randr"}, nil
  }

  return nil, fmt.Errorf("couldn't parse wlr-randr output")
}

func getSwaySize() (*TermSize, error) {
  cmd := exec.Command("swaymsg", "-t", "get_outputs")
  output, err := cmd.Output()
  if err != nil {
    return nil, fmt.Errorf("swaymsg error: %v", err)
  }

  // This is a simple parser - you might want to use proper JSON parsing
  re := regexp.MustCompile(`"current_mode":\s*{\s*"width":\s*(\d+),\s*"height":\s*(\d+)`)
  match := re.FindSubmatch(output)
  if match != nil {
    width, _ := strconv.Atoi(string(match[1]))
    height, _ := strconv.Atoi(string(match[2]))
    return &TermSize{Width: width, Height: height, Source: "sway"}, nil
  }

  return nil, fmt.Errorf("couldn't parse swaymsg output")
}

// GTK-based approach
func getGTKSize() (*TermSize, error) {
  // Try using zenity (GTK-based dialog)
  cmd := exec.Command("zenity", "--question", "--title=")
  // Start but don't wait
  if err := cmd.Start(); err != nil {
    return nil, fmt.Errorf("zenity error: %v", err)
  }

  // Give it a moment to create window
  time.Sleep(100 * time.Millisecond)

  // Get window ID and size
  windowID, err := exec.Command("xdotool", "search", "--name", "zenity").Output()
  if err != nil {
    cmd.Process.Kill()
    return nil, fmt.Errorf("xdotool error: %v", err)
  }

  // Kill zenity window
  defer cmd.Process.Kill()

  // Get geometry
  geom, err := exec.Command("xdotool", "getwindowgeometry", strings.TrimSpace(string(windowID))).Output()
  if err != nil {
    return nil, fmt.Errorf("xdotool geometry error: %v", err)
  }

  re := regexp.MustCompile(`Geometry: (\d+)x(\d+)`)
  match := re.FindSubmatch(geom)
  if match != nil {
    width, _ := strconv.Atoi(string(match[1]))
    height, _ := strconv.Atoi(string(match[2]))
    return &TermSize{Width: width, Height: height, Source: "gtk-zenity"}, nil
  }

  return nil, fmt.Errorf("couldn't parse zenity window geometry")
}

// Qt-based approach
func getQtSize() (*TermSize, error) {
  // Try using kdialog (Qt-based)
  cmd := exec.Command("kdialog", "--title=")
  if err := cmd.Start(); err != nil {
    return nil, fmt.Errorf("kdialog error: %v", err)
  }

  time.Sleep(100 * time.Millisecond)

  // Similar to GTK approach
  windowID, err := exec.Command("xdotool", "search", "--name", "kdialog").Output()
  if err != nil {
    cmd.Process.Kill()
    return nil, fmt.Errorf("xdotool error: %v", err)
  }

  defer cmd.Process.Kill()

  geom, err := exec.Command("xdotool", "getwindowgeometry", strings.TrimSpace(string(windowID))).Output()
  if err != nil {
    return nil, fmt.Errorf("xdotool geometry error: %v", err)
  }

  re := regexp.MustCompile(`Geometry: (\d+)x(\d+)`)
  match := re.FindSubmatch(geom)
  if match != nil {
    width, _ := strconv.Atoi(string(match[1]))
    height, _ := strconv.Atoi(string(match[2]))
    return &TermSize{Width: width, Height: height, Source: "qt-kdialog"}, nil
  }

  return nil, fmt.Errorf("couldn't parse kdialog window geometry")
}

// X11-specific approaches (from previous version, now with source tracking)
func getXwinfoSize() (*TermSize, error) {
  cmd := exec.Command("xdotool", "getactivewindow")
  windowID, err := cmd.Output()
  if err != nil {
    return nil, fmt.Errorf("xdotool error: %v", err)
  }

  cmd = exec.Command("xwininfo", "-id", strings.TrimSpace(string(windowID)))
  output, err := cmd.Output()
  if err != nil {
    return nil, fmt.Errorf("xwininfo error: %v", err)
  }

  widthRe := regexp.MustCompile(`Width: (\d+)`)
  heightRe := regexp.MustCompile(`Height: (\d+)`)

  widthMatch := widthRe.FindSubmatch(output)
  heightMatch := heightRe.FindSubmatch(output)

  if widthMatch == nil || heightMatch == nil {
    return nil, fmt.Errorf("couldn't parse xwininfo output")
  }

  width, _ := strconv.Atoi(string(widthMatch[1]))
  height, _ := strconv.Atoi(string(heightMatch[1]))

  return &TermSize{Width: width, Height: height, Source: "xwininfo"}, nil
}

// Try using TIOCGWINSZ ioctl
func getIoctlSize() (*TermSize, error) {
  type winsize struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
  }

  ws := &winsize{}
  retCode, _, errno := syscall.Syscall(
    syscall.SYS_IOCTL,
    uintptr(syscall.Stdin),
    uintptr(syscall.TIOCGWINSZ),
    uintptr(unsafe.Pointer(ws)))

  if int(retCode) == -1 {
    return nil, fmt.Errorf("ioctl error: %v", errno)
  }

  if ws.Xpixel > 0 && ws.Ypixel > 0 {
    return &TermSize{
      Width:  int(ws.Xpixel),
      Height: int(ws.Ypixel),
      Source: "ioctl",
    }, nil
  }

  return nil, fmt.Errorf("ioctl returned zero size")
}

func getOscSeq() (*TermSize, error) {
  ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
  if err != nil {
    return nil, fmt.Errorf("osc didn't work: %v", err)
  }
  return &TermSize{
    Width:  int(ws.Xpixel),
    Height: int(ws.Ypixel),
    Source: "osc",
  }, nil
}

func check_device_dims() (width, height int) {
  // Order methods by reliability and speed
  methods := []func() (*TermSize, error){
    getIoctlSize,   // Try ioctl first as it's fastest
    getWaylandSize, // Then try Wayland-specific methods
    getXwinfoSize,  // Then X11 methods
    getGTKSize,     // Then toolkit-specific methods
    getQtSize,
    getOscSeq,
  }

  for _, method := range methods {
    size, err := method()
    if err == nil && size.Width > 0 && size.Height > 0 {
      return size.Width, size.Height
    }
  }

  return 0, 0
}
