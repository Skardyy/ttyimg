# ttyimg 🔥  
a simple cli tool for encoding images into iterm / kitty / sixel format  

https://github.com/user-attachments/assets/f8cdff2e-fbfe-486f-84ba-330570a9e4de

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

> The following are supported but requires `Libreoffice`  
> They will take longer, because the extra step of converting the doc into an image  
- [X] DOCX  
- [X] XLSX  
- [X] PDF  
- [X] PPTX  

> ![Note]  
> i am open for suggestions on other backends for the document types  
> Libreoffice was chosen for it being the only crossplatform one  
