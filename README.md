<h1 align="center">ttyimg</h1>  
<p align="center">A simple cli tool for encoding images into iterm / kitty / sixel format.</p> 
<div align="center">
    
[![Static Badge](https://img.shields.io/badge/go.dev-00ADD8?style=flat&logo=go&logoColor=00ADD8&label=find%20at&labelColor=15161b)](https://pkg.go.dev/github.com/Skardyy/ttyimg)
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
  -f string
        fallback to when no protocol is supported: kitty, iterm, sixel (default "none")
  -h int
        Resize height
  -m string
        the resize mode to use when resizing: Fit, Strech, Crop (default "Fit")
  -p string
        Force protocol: kitty, iterm, sixel (default "auto")
  -w int
        Resize width
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
