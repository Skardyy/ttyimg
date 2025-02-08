# ttyimg 🔥  
a simple cli tool for encoding images into iterm / kitty / sixel format  

## Installation 📦
```sh
go install github.com/Skardyy/ttyimg
```

## Usuage 💡  
```sh
Usage: ttyimg [options] <path_to_image>
  -f string
        fallback to when no protocol is supported: kitty, iterm, sixel (default "none")
  -h int
        Resize height
  -p string
        Force protocol: kitty, iterm, sixel (default "auto")
  -w int
        Resize width
```
