<h1 align="center">ttyimg</h1>  
<p align="center">A powerfull cli tool for encoding images into <br/>
<b> Iterm2 / Kitty / Sixel </b> </p> 
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
         Resize width: <number> (pixels) / <number>px / <number>c (cells) / <number>% (default: 80%)
  -h string
         Resize height: <number> (pixels) / <number>px / <number>c (cells) / <number>% (default: 60%)
  -m string
         the resize mode to use when resizing: Fit, Strech, Crop (default: Fit)
  -center bool
         rather or not to center align the image (default: true)
  -p string
         Force protocol: kitty, iterm, sixel (default: auto)
  -f string
         fallback to when no protocol is supported: kitty, iterm, sixel (default: sixel)
  -spx string
         <width>x<height> or <width>x<height>xForce. specify the size of the winodw in px for fallback / overwrite (default: 1920x1080)
  -sc string
         <width>x<height> or <width>x<height>xForce. specify the size of the winodw in cell for fallback / overwrite (default: 120x30)
  -scale string
         <float>x<float> scales the spx and sc, only usefull for centering in smaller portions of the screen (default: 1x1)
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

## App Logic  
* first queries the size of the screen using:  
    *  cells: `\x1b[18t`
    *  px: `\x1b[14t`
* if neither works it fallbacks to:  
    *  cells: term.GetSize(fd), uses win api / ioctl respectfully, shouldn't fail unless stderr is not the terminal.  
    *  px: ioctl / windows api. windows shouldn't fail, just not as accurate. ioctl only fails if stderr is not the temrinal.  

Those options e.g (spx, sc, scale) aren't really important for normal users.  
but can be very powerfull for power users trying to call the program in emulated environments, like neovim \ tmux.  

> using those values we can use sizes  
> like c (cells) and % for resizing the image  
> and even center the image  
