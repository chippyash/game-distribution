# Building

This guide covers building `gamedist` (the CLI) and `gamedistapp` (the desktop app) from source.

## Toolchain

- Go 1.22 or later. `go.mod` says `go 1.22.0`. Install Go from [go.dev/dl](https://go.dev/dl/). The Go packaged with
  Ubuntu 22.04 is too old to fetch the right toolchain. With `GOTOOLCHAIN=auto` (the default), Go then fetches the
  toolchain that the dependencies need.
- A C compiler, for cgo, which the desktop app needs.

## Linux packages

The build also needs `make` and `zip`. The desktop app needs a C compiler and the OpenGL, X11 and Wayland development
headers:

```
sudo apt install build-essential zip libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev
```

The desktop app also needs these libraries at run time: `libgl1`, `libwayland-client0` and `libx11-6`.

## Make targets

- `make build`: the release build. It runs `clean` first, which also runs `go mod tidy` and can rewrite `go.mod` and
  `go.sum`. It produces:
  - `bin/gamedist` and `bin/gamedist.exe`
  - `bin/gamedistapp` and `bin/gamedistapp.exe`
  - `dist/gamedist.zip`, which holds the four binaries and `README.md`
- `make build-gamedist`: the CLI only.
- `make build-desktop`: the desktop app for Linux and Windows.
- `make desktop-icon`: compiles the Windows icon resource.

`dist/` is git-ignored.

Do not run `make test`. It runs `go get -u`, which rewrites `go.mod`.

## Windows build

Cross-compiling the Windows desktop app needs the MinGW-w64 compiler and `windres`:

```
sudo apt install gcc-mingw-w64-x86-64
```

`make desktop-icon` compiles `src/desktop/resources/icon-windows.rc` into `src/desktop/icon_windows_amd64.syso`, which Go
links into the Windows executable. The `.syso` file is git-ignored. Run `make desktop-icon` again if it is missing.

The Windows build uses `-H=windowsgui`, so the app starts without a console window.

## Testing the Windows build on Linux

Wine can run the Windows app for a quick check. Use 64-bit Wine (`wine64`) with a separate prefix, so the default Wine
setup is untouched:

```
WINEPREFIX=<dir> WINEDEBUG=-all WINEDLLOVERRIDES="mscoree,mshtml=" wine64 bin/gamedistapp.exe
```

Wine is not a real Windows install. The final check of any Windows behaviour needs a Windows machine.

## Version

`VERSION` in the repository root is injected into both binaries at build time. Edit that file to change the version.
