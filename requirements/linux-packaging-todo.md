# Linux Packaging Implementation To-Do

This is the implementation plan for `requirements/linux-packaging.md`. Each task is marked with who carries it out:

- **[Claude]**: I can do this in the repository.
- **[You]**: needs an action from you, such as supplying a file, installing software, or pushing and releasing.
- **[Both]**: I do the work, and you check or decide something.

Status: not started. Work is on the branch `feature/linux_packaging`, created from `master`.

## Decisions this plan relies on

These come from the answers in `requirements/linux-packaging.md`:

- `nfpm` builds both the `.deb` and the `.rpm` from one YAML file. It is downloaded as a pinned release, and its SHA-256 checksum is checked.
- The `.deb` is checked on `ubuntu-22.04` and `debian:12`. The `.rpm` is checked on a pinned `fedora:<N>` container.
- Packages are x86_64 only.
- The package version is `VERSION` without the leading `v`.
- Tags with a hyphen build no Linux packages.
- The `.tar.gz` has no `.desktop` file. The `.deb` and `.rpm` do.

## Phase 0: Prerequisites

- [x] **[You]** Supply the new desktop entry icon. Save it at `src/desktop/resources/` and tell me the file name. It is needed in Phase 2.
- [x] **[You]** Confirm the licence text for `LICENSE`. Use the text of `LICENSE.md`. Done: `LICENSE` is a copy of `LICENSE.md`, which the source headers and the README already refer to as "the LICENSE file".
  - [x] **[You]** Choose how to test the release workflow before the first real release. Decided: a `workflow_dispatch` trigger on the branch. Test tags with a hyphen build o packages, so a tag cannot test them.
- [x] **[You]** Install the tools for local checks: `desktop-file-utils` (for `desktop-file-validate`), and `podman` or `docker` (for the Debian and Fedora containers). Checked: `desktop-file-validate` 0.28, `podman` and `docker` are installed and working, and the user can reach the Docker daemon.
- [x] **[Claude]** Pin the `nfpm` version and record its SHA-256 checksum for the x86_64 Linux release. Done: `nfpm` v2.48.0, published 2026-10-12, checksum `6d7f9a28…341750`, matched against the release's `checksums.txt`. Recorded in the workflow's `env` block. `docs/building.md` gets the same values in Phase 8.
- [x] **[Claude]** Pin the Fedora release for the `.rpm` check and record it in one place, so it is easy to bump. Done: Fedora 45, the current stable release (Fedora 46 is Rawhide). Recorded as `FEDORA_VERSION` in the workflow's `env` block.

## Phase 1: Licence and version

- [x] **[Claude]** Add `LICENSE` at the repository root, once Phase 0 settles its text. Done: a copy of `LICENSE.md`, which is still in the repository (see Phase 0).
- [x] **[Claude]** Add a `make` helper or shell line that gives the package version: `VERSION` with the leading `v` removed. Done: `PKG_VERSION` in the `Makefile`, read from the `VERSION` file. It gives `1.2.1` for `v1.2.1`. The hyphen rule for tags is applied in the workflow (Phase 6), not here.

## Phase 2: Desktop entry and icon assets

- [x] **[Claude]** Add `packaging/gamedistapp.desktop` with `Name=Game Distributor`, a short `Comment`, `Exec=gamedistapp`, `Icon=gamedistapp`, `Terminal=false`, `Type=Application` and `Categories=Utility;`.
- [x] **[Claude]** Install the icon under the standard `hicolor` path, using the sizes from the supplied file. Use the same icon name as the `Icon=` line. Convert it to PNG if the supplied file is not one, as the requirements say. Done: PNG files at 48, 128 and 256 pixels in `packaging/icons/hicolor/<size>/apps/gamedistapp.png`, laid out as the hicolor theme expects. The `nfpm` config in Phase 3 installs them.
- [x] **[Claude]** Run `desktop-file-validate` on the desktop entry locally. Passes.

## Phase 3: nfpm configuration and local build

- [x] **[Claude]** Add `packaging/nfpm.yaml` for the `.deb` and `.rpm`. It lists the files, the install paths (`/usr/bin`, `/usr/share/applications`, the icon path, `/usr/share/doc/gamedist`), the `libgl1`, `libwayland-client0` and `libx11-6` dependencies, and their RPM names. Done: `packaging/nfpm.yaml` with the files, `Utility` desktop entry, the `.deb` and `.rpm` dependencies, and the install and remove scripts in `packaging/scripts/`. The maintainer line has no email address, so it does not publish one.
- [x] **[Claude]** Find the RPM names of the three runtime libraries on a Fedora container, and record them in the file. The requirements say they differ from the Debian names. Done: `libglvnd-glx` (for `libGL.so.1`), `libwayland-client` and `libX11`, found with `dnf provides` on `fedora:45`. They are in the `rpm` override of `nfpm.yaml`.
- [x] **[Claude]** Add a `make package-linux` target. It requires `bin/` from `make build`, takes the version from `VERSION`, and writes `dist/gamedist_<version>-1_amd64.deb`, `dist/gamedist-<version>-1.x86_64.rpm` and `dist/gamedist-<version>-linux-amd64.tar.gz`. Done: `make package-linux` builds the `.deb`, `.rpm` and `.tar.gz` into `dist/`. Built locally from the existing binaries. The `.deb` is named `gamedist_1.2.1-1_amd64.deb`, because nfpm adds the default revision `-1`. Decided: keep nfpm's name. Phase 8 updates the README example and the requirements to `gamedist_<version>-1_amd64.deb`.
- [x] **[Claude]** Make the target fail with a clear message if `nfpm` is missing, and say which version to install. Done: a missing `nfpm` and missing binaries each stop the target with a message. Both were tested.
- [x] **[Claude]** Confirm `make build` and `dist/gamedist.zip` are unchanged. The Linux packages are separate files, not new contents of the zip. Checked: the `build-gamedist` and `build-desktop` recipes are unchanged, and both ran, so the zip still holds the same four binaries and `README.md`. The full `make build` was not run, because it runs `clean`, which runs `go mod tidy`. `go.mod` is unchanged.

## Phase 4: The `.tar.gz`

- [x] **[Claude]** Build the `.tar.gz` with one top-level directory `gamedist-<version>`, holding `gamedist`, `gamedistapp`, `README.md` and `LICENSE`. No `.desktop` file, as decided. Done: `make package-linux` builds `dist/gamedist-1.2.1-linux-amd64.tar.gz`. It holds `gamedist`, `gamedistapp`, `README.md` and `LICENSE`, with no `.desktop` file.
- [x] **[Claude]** Check with `tar -tzf` that nothing is outside that directory. Done: every entry is under `gamedist-1.2.1/`, and none is outside it.

## Phase 5: Local verification

- [x] **[Claude]** Build the `.deb` and `.rpm` locally, and inspect their file lists with `dpkg -c` and `rpm -qlp`. Done: both built. The file lists were checked inside the Ubuntu and Fedora containers, since the host has no rpm database.
- [x] **[Claude]** Compare the library dependencies with `ldd bin/gamedistapp`. The earlier check showed `libGL`, `libwayland-client` and `libX11`, with the rest coming in through them. Done: `ldd` shows 11 shared libraries. Every one comes from `libgl1`, `libwayland-client0`, `libx11-6` (Debian) or `libglvnd-glx`, `libwayland-client`, `libX11` (RPM), and both installs report no missing library.
- [x] **[Claude]** Install the `.deb` in an `ubuntu-22.04` container and in a `debian:12` container. Run `gamedist help` after each install, and check the desktop entry and icon are present. Earlier, with a binary built on this host, the desktop app failed on Ubuntu 22.04 and Debian 12 with `GLIBC_2.38`. Done on Ubuntu 22.04: `gamedistapp` was built in a 22.04 container with Go 1.27.3. It needs at most `GLIBC_2.34`, has no missing libraries with the declared runtime packages installed, and stays running under Xvfb for 8 seconds. Also done on Debian 12 (glibc 2.36) with the same build: no missing libraries, and it stays running for 8 seconds. Still open: installing the `.deb` built from the 22.04 binaries, which the release workflow will do. Done with the packages built from the 22.04 binary: the `.deb` installs on Ubuntu 22.04 and Debian 12, `gamedist help` exits 0, the desktop app starts and stays running for 8 seconds under Xvfb with no missing libraries, and the desktop entry and icons are installed. The `.rpm` installs and runs the same way on Fedora 45.
- [x] **[Claude]** Install the `.rpm` in a `fedora:<N>` container with `dnf install --nogpgcheck`, then run `gamedist help`. Done on `fedora:45` with `dnf install --nogpgcheck`: `gamedist help` exits 0, the desktop entry validates, and no library is missing. Removing the package also removed the empty `/usr/share/doc/gamedist` folder, which led to a fix: the folder is now a `dir` entry in `nfpm.yaml`.
- [x] **[Claude]** Remove each package and confirm that every file it installed is gone, while the runtime libraries stay. Done: the `.deb` and `.rpm` remove every file the package installed, and the `.rpm` leaves no empty folder. Library behaviour follows the decision in `linux-packaging.md`: the `.deb` keeps its libraries, and the `.rpm` keeps a library that was already installed but removes one that gamedist brought in.
- [x] **[Claude]** Confirm the desktop database and icon cache refreshes do nothing harmful when their tools are missing. Done: the `.deb` installs on Ubuntu 22.04 with the tools present, and the icon cache is written. The `.deb` on Debian 12 and the `.rpm` on Fedora 45 install with the tools absent, and both exit 0.
- [x] **[You]** Look at the desktop entry on a real Linux desktop: it should appear under Utilities after the install.

## Phase 6: Release workflow (`.github/workflows/fsl-release.yml`)

- [x] **[Claude]** Add a step that downloads the pinned `nfpm` and checks its SHA-256 checksum. The step fails if the checksum does not match. Done: the `Install nfpm` step downloads `NFPM_VERSION` and runs `sha256sum -c`, which fails on a mismatch.
- [x] **[Claude]** Add the packaging step after `make build`. It builds the packages from `bin/`, and it is skipped for tags with a hyphen. Done: the `Decide What to Build` step sets `packages=false` for a tag with a hyphen, and the packaging steps are skipped then.
- [x] **[Claude]** Add a `.deb` check job on `ubuntu-22.04` and a second one in a `debian:12` container. Each installs the package and runs `gamedist help`. Done: `check-deb-ubuntu` (`ubuntu:22.04`) and `check-deb-debian` (`debian:12`) install the `.deb` and run `packaging/scripts/check-install.sh`, which runs `gamedist help`.
- [x] **[Claude]** Add an `.rpm` check job in a pinned `fedora:<N>` container, using `dnf install --nogpgcheck`. Done: `check-rpm-fedora` runs in `fedora:` with the version from `FEDORA_VERSION`, using `dnf install --nogpgcheck`, and runs the same check script.
- [x] **[Claude]** Add `desktop-file-validate` to the checks, and confirm the installed entry exists. Done: `check-install.sh` runs `desktop-file-validate`, checks the entry and the three icons, and fails on a missing library.
- [x] **[Claude]** Add a `workflow_dispatch` trigger, so the packaging and checks can run on the branch without a tag. A manual run builds and checks the packages, but does not create a tag or a release, so it leaves nothing on GitHub. Done: `workflow_dispatch` runs the build and checks, and the `draft-release` job runs for tags only.
- [x] **[Claude]** Make the release creation depend on the checks, so a failed check means no draft release. Done: `draft-release` needs the three checks, and runs only when they pass or when the tag has a hyphen and builds no packages.
- [x] **[Claude]** Attach the `.deb`, `.rpm` and `.tar.gz` with the zip, using the existing `softprops/action-gh-release` step. Done: the `packages` artifact holds the zip and the packages, and `softprops/action-gh-release` attaches `release/*`.
- [x] **[Claude]** Check that the release notes step and the draft release behaviour are unchanged. Done: the release notes step is unchanged. It now runs in the `draft-release` job, with the same dates and the same `git describe` range. `docs/releasing.md` still describes the old steps, and Phase 8 updates it.

## Phase 7: Testing the workflow

- [x] **[You]** Run the workflow from the branch with `workflow_dispatch`, and tell me the result. Run `42229341957` on `feature/linux_packaging`: `build`, `check-deb-ubuntu`, `check-deb-debian` and `check-rpm-fedora` passed. `draft-release` was skipped, as intended for a manual run. No tag or release was created.
- [x] **[Claude]** Read the run log and fix anything that fails. Repeat until the checks pass. Nothing failed, so no fix was needed.
- [x] **[You]** Delete any test release and tag afterwards, as `docs/releasing.md` describes. Tell me when that is done.

## Phase 8: Documentation

- [x] **[Claude]** `docs/building.md`: add the Linux packages section, the packaging tools with their install commands, the pinned `nfpm` version, and the `make package-linux` target. Done: a "Linux packages" section with the nfpm install commands and checksum, and the `make package-linux` target in the list of targets.
- [x] **[Claude]** `docs/releasing.md`: list the new artifacts in the workflow steps, name the pinned Fedora release and say how to bump it, and describe the hyphen-tag rule. Done: the workflow jobs, the pinned values and how to bump them, the hyphen-tag rule, the manual run, and the package names.
- [x] **[Claude]** `README.md`: add the `.deb` and `.rpm` install lines, using `gamedist_<version>-1_amd64.deb` for the `.deb` name, the `.rpm` `--nogpgcheck` note, and "Fedora N and later" with the pinned N. Done: package install lines for Debian, Ubuntu and Fedora, the `--nogpgcheck` note, and the supported versions. The manual steps follow under "Manual install".
- [x] **[Claude]** Check that the README links still work on GitHub once the files are on `master`. Partly done: the relative links and anchors in `docs/building.md` and `docs/releasing.md` resolve. The README uses absolute GitHub links on `master`, which can only be checked after the branch is merged.

## Phase 9: Review and release

- [ ] **[Claude]** Run `make lint` and `go vet` with the `desktop` tag. Run `go test ./src/...` as `CONTRIBUTING.md` describes. Do not run `make test`.
- [ ] **[Claude]** Check the requirements file against the result, and mark each requirement that is met.
- [ ] **[You]** Review the diff and decide on the commit. Commit only when you ask.
- [ ] **[You]** Push the branch and open the pull request. Merge and tag the release when you are ready.
- [ ] **[You]** After the release, check the attached packages on a real Ubuntu, Fedora and Debian machine if you have one.
