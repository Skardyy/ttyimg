package main

import (
  "fmt"
  "os"
  "path/filepath"
  "sync"
  "time"
)

func get_log_path() string {
  exePath, _ := os.Executable()
  exeDir := filepath.Dir(exePath)
  return filepath.Join(exeDir, "logs.log")
}

type Logger struct {
  logFile *os.File
  mu      sync.Mutex
  logPath string
  enabled bool
}

func (l *Logger) Init(path string, enabled bool) error {
  l.enabled = enabled
  l.logPath = path
  if !l.enabled {
    return nil
  }

  dir := filepath.Dir(path)
  if _, err := os.Stat(dir); os.IsNotExist(err) {
    if err := os.MkdirAll(dir, 0755); err != nil {
      return fmt.Errorf("failed to create log directory: %v", err)
    }
  }

  f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
  if err != nil {
    return fmt.Errorf("failed to open log file: %v", err)
  }

  l.logFile = f

  return nil
}

func (l *Logger) Write(message string) {
  if !l.enabled {
    return
  }
  l.mu.Lock()
  defer l.mu.Unlock()

  if message != "" {
    timestamp := time.Now().Format("02/01/2006 15:04:05")
    fmt.Fprintf(l.logFile, "<%s>:\n    %s\n\n", timestamp, message)
  }
}

func (l *Logger) Close() {
  if !l.enabled {
    return
  }
  if l.logFile != nil {
    err := l.logFile.Close()
    if err != nil {
      fmt.Fprintln(os.Stderr, "Error closing log file:", err)
    }
  }
}
