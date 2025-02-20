<h1 align="center">ttyimg</h1>  
<p align="center">🖼️ A simple cli tool for encoding images into <b>Iterm2 / Kitty / Sixel</b> 🖼️</p> 
<div align="center">
    

[![Static Badge](https://img.shields.io/badge/go.dev-1e2029?style=flat&logo=go&logoColor=00ADD8&label=find%20at&labelColor=15161b)](https://pkg.go.dev/github.com/Skardyy/ttyimg) ˙ ![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/Skardyy/ttyimg/release.yml?style=flat&labelColor=15161b&color=1e2029)


</div>

---
https://github.com/user-attachments/assets/92b635e9-7ffe-4eed-8abe-a5d593504990

## Installation 📦
```sh
go install github.com/Skardyy/ttyimg@latest
```

## Usuage 💡  
```sh
Usage: ttyimg [options] <path_to_image>
  -w string
         Resize width: 100 (pixels) / 100px / 100c (cells) / 100% (default: 0)
  -h string
         Resize height: 100 (pixels) / 100px / 100c (cells) / 100% (default: 0)
  -m string
         the resize mode to use when resizing: Fit, Strech, Crop (default: Fit)
  -center bool
         rather or not to center align the image (default: false)
  -p string
         Force protocol: kitty, iterm, sixel (default: auto)
  -f string
         fallback to when no protocol is supported: kitty, iterm, sixel (default: sixel)
  -screen string
         what to use as fallback if the app fails to query the size by itself (default: 1920x1080)
  -forceScreen bool
         rather or not to force the screen size and not attempt to query (default: false)
  -cache bool
         rather or not to cache the heavy operations (default: true)
```

## Supports ✨  
- [X] PNG  
- [X] JPEG  
- [X] TIFF  
- [X] SVG  
- [X] WEBP  
- [X] DOCX  
- [X] XLSX  
- [X] PDF  
- [X] PPTX  

> DOCX, XLSX, PDF and PPTX require
><details>
>  <summary>Libreoffice</summary>
> 
>  ```txt
>    make sure its installed and in your path  
>    * windows: in windows its called soffice and should be in C:\Program Files\LibreOffice\program 
>    * linux: should add it to path automatically
>  ```
> </details>

> [!Note]  
> i am open for suggestions on other backends for the document types  
> Libreoffice was chosen for it being the only crossplatform one  
