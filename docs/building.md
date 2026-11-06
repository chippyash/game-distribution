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

## Linux packages

`make package-linux` builds the `.deb`, the `.rpm` and the `.tar.gz` from `bin/`, so run `make build` first. It writes them to
`dist/`. It needs [nfpm](https://nfpm.goreleaser.com) on your `PATH`, at the version pinned in the release workflow, which is
v2.48.0 at the time of writing:

```
curl -sSfL -o /tmp/nfpm.tar.gz https://github.com/goreleaser/nfpm/releases/download/v2.48.0/nfpm_2.48.0_Linux_x86_64.tar.gz
echo "6d7f9a28b2f397934403950fcbba83f598b08d02edbe79a0815b3b4897341750  /tmp/nfpm.tar.gz" | sha256sum -c -
sudo tar -C /usr/local/bin -xzf /tmp/nfpm.tar.gz nfpm
```

The checksum is the one in the release's `checksums.txt`. The `.rpm` is built without any RPM tools, and the `.deb` without
`dpkg-deb`, so nothing else is needed. The packages contain the binaries, the desktop entry and icons, `README.md` and
`LICENSE`. The desktop entry and icons are in `packaging/`, and `packaging/nfpm.yaml` describes the packages.

The release workflow builds the packages on `ubuntu-22.04` and checks each one in a clean container. See
[Releasing](releasing.md).

## Make targets

- `make build`: the release build. It runs `clean` first, which also runs `go mod tidy` and can rewrite `go.mod` and
  `go.sum`. It produces:
  - `bin/gamedist` and `bin/gamedist.exe`
  - `bin/gamedistapp` and `bin/gamedistapp.exe`
  - `dist/gamedist.zip`, which holds the four binaries and `README.md`
- `make build-gamedist`: the CLI only.
- `make build-desktop`: the desktop app for Linux and Windows.
- `make desktop-icon`: compiles the Windows icon resource.
- `make package-linux`: the Linux `.deb`, `.rpm` and `.tar.gz` packages. See [Linux packages](#linux-packages).

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

## macOS

The desktop app cannot be built for macOS on Linux. It needs Apple's SDK and a macOS compiler, and no macOS build is
provided with the release.

## Testing the Windows build on Linux

Wine can run the Windows app for a quick check. Use 64-bit Wine (`wine64`) with a separate prefix, so the default Wine
setup is untouched:

```
WINEPREFIX=<dir> WINEDEBUG=-all WINEDLLOVERRIDES="mscoree,mshtml=" wine64 bin/gamedistapp.exe
```

Wine is not a real Windows install. The final check of any Windows behaviour needs a Windows machine.

## Version

`VERSION` in the repository root is injected into both binaries at build time. Edit that file to change the version.
