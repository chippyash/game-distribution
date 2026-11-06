# Linux Packaging Requirements

## Overview

Today the Linux binaries reach users inside `gamedist-<tag>.zip`, next to the Windows binaries and `README.md`. The user
copies `gamedist` and `gamedistapp` to a directory on their `PATH` by hand, and installs the desktop app's runtime
libraries with `apt` as a separate step. This plan adds Linux packages that install the CLI and the desktop app in one
step, and declare the libraries the desktop app needs, so that the desktop app starts on a supported distribution without
the user looking up the library names.

## Scope

- Packaging only. No change to the CLI or desktop app behaviour.
- The packages are built from the same binaries that `make build` produces. The packaging step must not compile the code
  a second time with different flags.
- The Windows build, the release zip and the Makefile targets that already exist stay as they are. The Linux packages are
  added to the release, next to the zip, not in place of it.
- The packages are built on `ubuntu-22.04` in the release workflow, so they keep the glibc 2.35 baseline described in
  [docs/releasing.md](../docs/releasing.md).

## Candidate Formats

| Format | Covers | Notes |
|---|---|---|
| `.deb` | Debian and Ubuntu, and derivatives | Declares the runtime libraries as dependencies, so `apt` installs them. Installs to `/usr/bin` |
| `.rpm` | Fedora, openSUSE and other RPM distributions | Same role as `.deb`. Needs an RPM build tool on the runner |
| `.tar.gz` | Any Linux | No install step. Holds the binaries, `README.md` and a licence file, and the user copies them to a path |
| AppImage | Any Linux with FUSE | One file for the desktop app. Does not cover the CLI well, so it is probably not needed |
| Flatpak or Snap | Sandboxed desktop installs | Larger effort, with sandbox permissions to justify. Not planned for the first release |

The first release should cover `.deb`, `.rpm` and a `.tar.gz`. Other formats are added only if users ask for them.

## Functional Requirements

- `make` provides a target that builds the Linux packages from `bin/`, for example `make package-linux`. It writes the
  output to `dist/`, and `dist/` stays git-ignored.
- The `.deb` package installs `gamedist` and `gamedistapp` to `/usr/bin`, and installs a desktop entry and an icon for
  `gamedistapp`, so that it appears in the application menu.
- The desktop entry for `gamedistapp` is in the freedesktop `Utility` category (`Categories=Utility;`), so desktop
  environments list it under Utilities, or the nearest equivalent. It has the name "Game Distributor", a short comment
  that describes what the app does, and the icon. It is installed to `/usr/share/applications/`.
- The `gamedist` CLI is run from a terminal, so it gets no menu entry. It is installed to `/usr/bin` only.
- The package refreshes the desktop database and the icon cache in its install and remove scripts, so the entry appears
  without a logout. The `.deb` and `.rpm` scripts use the standard `update-desktop-database` and `gtk-update-icon-cache`
  tools, and do nothing if those tools are absent.
- The `.deb` package declares the runtime libraries the desktop app needs as dependencies: `libgl1`, `libwayland-client0`
  and `libx11-6`, as listed in [docs/building.md](../docs/building.md). Installing the package installs them.
- The `.deb` package includes `README.md` and `LICENSE` under `/usr/share/doc/gamedist/`. The repository has no `LICENSE`
  file yet, although the README refers to one. Adding it to the repository is part of this work.
- The `.deb` package name is `gamedist`. Its version is `VERSION` with the leading `v` removed, so `v1.2.1` becomes `1.2.1`.
- The `.rpm` package has the same content and install layout as the `.deb`: `gamedist` and `gamedistapp` in `/usr/bin`,
  the desktop entry in the `Utility` category with its icon, and `README.md` and `LICENSE` under `/usr/share/doc/gamedist/`.
- The `.rpm` package recommends the same runtime libraries, under their RPM names, with `Recommends:` and not `Requires:`.
  dnf installs recommended packages by default. The names are checked against a Fedora image before the package is first
  released, because they differ from the Debian names.
- The `.rpm` package has the same version as the `.deb`, `VERSION` with the leading `v` removed, with the RPM release field
  set to `1`. Its file name follows the RPM convention, for example `gamedist-<version>-1.x86_64.rpm`.
- The `.tar.gz` package has one top-level directory named `gamedist-<version>`. It holds `gamedist`, `gamedistapp`,
  `README.md` and `LICENSE`. Extracting it creates nothing outside that directory. It has no install step, so it adds no menu
  entry. The README says how to add one by hand, or the user can use the `.deb` or `.rpm`.
- The release workflow attaches the Linux packages to the same release as `gamedist-<tag>.zip`. The workflow still creates
  a draft release, as it does now.
- The package build needs no network access beyond the package sources that the workflow already uses, except for
  downloading the pinned `nfpm` release. That download is checked against its SHA-256 checksum before it is used.
- Each package is checked in the release workflow before it is attached. The check installs the `.deb` on a clean
  `ubuntu-22.04` image and again on a clean `debian:12` container, and the `.rpm` on a clean `fedora:<N>` container. It runs
  `gamedist help` after each install. `<N>` is pinned in the workflow and bumped by hand at each release. Running `gamedistapp` needs a display, so the check only confirms that the binary starts with
  `--help` or exits cleanly, if the app supports that. If it does not, the check confirms that the binary is present and
  executable. A failed check fails the workflow, so no draft release is created.
- The workflow downloads a pinned release of `nfpm` and checks its SHA-256 checksum before it uses the tool.
- Tags that contain a hyphen, such as `v0.0.1-test`, build the zip only and no Linux packages. The release workflow still
  creates the draft release for them, as it does now.
- The packages install only their own files. The only other changes are the standard desktop database and icon cache
  refreshes described above, and the packages do not edit the user's shell profile.
- Removing the `.deb` or `.rpm` package removes every file it installed, including the desktop entry and icon.
- Removing the `.deb` leaves the runtime libraries in place, as apt does. Removing the `.rpm` leaves a library in place if it
  was installed before gamedist. dnf removes a library that gamedist brought in, because nothing else needs it. This is
  dnf's default behaviour, and the package does not change it.
- The check in the release workflow also confirms that the installed desktop entry exists and passes
  `desktop-file-validate`.

## Non-Functional Requirements

- Build with the same toolchain and the same `-ldflags` as the existing targets, so both binaries report the same
  version in every format.
- The packages are built for x86_64 only. Other architectures are out of scope for now.
- Keep the packaging scripts in the repository, and keep them short enough to read in one sitting. `nfpm` is the only
  packaging tool added.
- No new runtime dependency is added to either binary.
- Use British spelling, as in the existing documentation.

## Documentation

- `docs/building.md` gains a section on the Linux packages, the packaging tools and their install commands on the build
  machine, and the `make` target that builds them.
- `docs/releasing.md` lists the new artifacts in the workflow steps, names the pinned Fedora release and says how to bump
  it, and updates the release steps and the test release cleanup if the workflow changes.
- `README.md` installation section for Linux names the `.deb` and `.rpm` as the simplest routes, and keeps the zip and
  `.tar.gz` routes. Each line says how to install its package, for example `sudo apt install ./gamedist_<version>-1_amd64.deb`
  or `sudo dnf install ./gamedist-<version>-1.x86_64.rpm`. The `.rpm` is not signed, so Fedora refuses it unless the user
  adds `--nogpgcheck`, and the README says so. The install line names Fedora N and later, where N is the Fedora release
  pinned in the workflow.

## Open Questions

- **Decided:** the first release includes `.rpm`, with `.deb` and `.tar.gz`. This question is closed.
- Should the `.deb` package be signed, or published to an apt repository? Proposed: no. Attach the file to the release,
  as the Windows binaries are. The `.deb` is then installed with `apt install ./file.deb`, which does not need a
  repository or a signing key. **Answered:** no signing is required for now.
- Which tool builds the `.deb` and `.rpm`? Options: `dpkg-deb` and `rpmbuild`, one native tool per format, or `nfpm`,
  which builds both from one YAML description. **Answered:** use `nfpm`, with the native tools as a fallback if `nfpm`
  cannot express what the packages need.
- Should the desktop entry use the Windows icon, or a PNG icon? The repository has `src/desktop/resources/icon-linux.jpeg`
  but no PNG. **Answered:** the user will supply a new desktop entry icon as the base for the packages. **Pending:** the
  icon file has not yet been added to the repository, so the icon sizes and path are not set.
- What minimum Debian and Ubuntu versions should the `.deb` target? **Answered:** the proposal is accepted, Ubuntu 22.04
  and later, and Debian 12 and later.
- Should the package check run on every tag, or only on a manual trigger? **Answered:** every tag.
- Should the CLI also get a menu entry, for example one that opens a terminal running `gamedist help`? **Answered:** no.

## Decisions

Answers to the review questions:

- The unsigned `.rpm` is documented with `--nogpgcheck`. Signing is not planned for now.
- The `.rpm` check runs on a pinned `fedora:<N>` container. N is bumped by hand at each release, and the docs say "Fedora N and later".
- The `.deb` is also checked on a `debian:12` container, so the Debian 12 claim is tested.
- Packages are x86_64 only.
- The package version drops the leading `v` from the tag.
- A failed package check fails the workflow, so no draft release is created.
- The `LICENSE` file is added to the repository as part of this work, and is included in every package.
- `nfpm` is downloaded as a pinned release with its SHA-256 checksum checked.
- Tags with a hyphen, such as test tags, build no Linux packages.
- The `.tar.gz` does not include a `.desktop` file. The README explains how to add a menu entry by hand.
- The icon is handled during implementation, once the file has been supplied.
- The `.rpm` uses `Recommends:` for the runtime libraries. dnf then keeps libraries that were installed before gamedist, but
  removes those that gamedist brought in. Keeping the libraries on removal would need a dnf-specific scriptlet, which
  hung when tested, so it is not used.

## Status

Checked at the end of implementation (Phase 9). "Met" means verified locally or in the dispatch run. "Met in the workflow"
means the workflow does it, but no tag run has exercised it yet.

- Met: `make package-linux` builds the `.deb`, `.rpm` and `.tar.gz` from `bin/`, and fails clearly without `nfpm` or binaries.
- Met: the `.deb` and `.rpm` install `gamedist` and `gamedistapp` to `/usr/bin`, with the desktop entry, icons, `README.md`
  and `LICENSE`. The entry is in the `Utility` category and passes `desktop-file-validate`.
- Met: the CLI has no menu entry, and the `.tar.gz` has no `.desktop` file.
- Met: removing each package removes every file it installed. Library behaviour is as amended above.
- Met: the desktop database and icon cache refreshes do nothing harmful when their tools are missing.
- Met: the `.deb` declares its runtime libraries, and the `.rpm` recommends them, as amended above.
- Met: the `.deb` is named `gamedist_<version>-1_amd64.deb`, and the `.rpm` is named `gamedist-<version>-1.x86_64.rpm`.
- Met: the `.deb` installs and runs on Ubuntu 22.04 and Debian 12, and the `.rpm` on Fedora 45. The desktop app runs on each
  when built in a 22.04 container.
- Met in the workflow: each package is checked in a clean container before it is attached, and a failed check fails the
  workflow and creates no draft release. The dispatch run passed all checks.
- Met in the workflow: `workflow_dispatch` runs the build and checks without a tag or release.
- Met in the workflow: a tag with a hyphen builds the zip only.
- Met in the workflow: the packages and the zip are attached to the draft release. Not yet exercised on a tag.
- Met: the package build needs no network access beyond the existing package sources, except the pinned, checksum-checked
  `nfpm` download, as amended above.
- Not yet checked: the packages on a real Ubuntu, Fedora and Debian desktop (see the plan's Phase 9).
