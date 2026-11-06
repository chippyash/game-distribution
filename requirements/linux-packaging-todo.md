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

- [ ] **[You]** Supply the new desktop entry icon. Save it at `src/desktop/resources/` and tell me the file name. It is needed in Phase 2.
- [ ] **[You]** Confirm the licence text for `LICENSE`. The README refers to FSL-1.1, so the file should be the Functional Source License 1.1 text, with the copyright line used in the header of the source files. Tell me if you want it copied from the FSL site or pasted in by you.
- [ ] **[You]** Choose how to test the release workflow before the first real release. Test tags with a hyphen now build no packages, so a tag cannot test them. Options: a `workflow_dispatch` trigger on the branch, or a temporary non-hyphen test tag that is deleted afterwards. Needed in Phase 7.
- [ ] **[You]** Install the tools for local checks: `desktop-file-utils` (for `desktop-file-validate`), and `podman` or `docker` (for the Debian and Fedora containers). Claude can check which are present.
- [ ] **[Claude]** Pin the `nfpm` version and record its SHA-256 checksum for the x86_64 Linux release. Record both in the workflow and in `docs/building.md`.
- [ ] **[Claude]** Pin the Fedora release for the `.rpm` check and record it in one place, so it is easy to bump.

## Phase 1: Licence and version

- [ ] **[Claude]** Add `LICENSE` at the repository root, once Phase 0 settles its text.
- [ ] **[Claude]** Add a `make` helper or shell line that gives the package version: `VERSION` with the leading `v` removed. Test it with `v1.2.1` and with a hyphen tag, to confirm the hyphen case is caught.

## Phase 2: Desktop entry and icon assets

- [ ] **[Claude]** Add `packaging/gamedistapp.desktop` with `Name=Game Distributor`, a short `Comment`, `Exec=gamedistapp`, `Icon=gamedistapp`, `Terminal=false`, `Type=Application` and `Categories=Utility;`.
- [ ] **[Claude]** Install the icon under the standard `hicolor` path, using the sizes from the supplied file. Use the same icon name as the `Icon=` line. Convert it to PNG if the supplied file is not one, as the requirements say.
- [ ] **[Claude]** Run `desktop-file-validate` on the desktop entry locally.

## Phase 3: nfpm configuration and local build

- [ ] **[Claude]** Add `packaging/nfpm.yaml` for the `.deb` and `.rpm`. It lists the files, the install paths (`/usr/bin`, `/usr/share/applications`, the icon path, `/usr/share/doc/gamedist`), the `libgl1`, `libwayland-client0` and `libx11-6` dependencies, and their RPM names.
- [ ] **[Claude]** Find the RPM names of the three runtime libraries on a Fedora container, and record them in the file. The requirements say they differ from the Debian names.
- [ ] **[Claude]** Add a `make package-linux` target. It requires `bin/` from `make build`, takes the version from `VERSION`, and writes `dist/gamedist_<version>_amd64.deb`, `dist/gamedist-<version>-1.x86_64.rpm` and `dist/gamedist-<version>-linux-amd64.tar.gz`.
- [ ] **[Claude]** Make the target fail with a clear message if `nfpm` is missing, and say which version to install.
- [ ] **[Claude]** Confirm `make build` and `dist/gamedist.zip` are unchanged. The Linux packages are separate files, not new contents of the zip.

## Phase 4: The `.tar.gz`

- [ ] **[Claude]** Build the `.tar.gz` with one top-level directory `gamedist-<version>`, holding `gamedist`, `gamedistapp`, `README.md` and `LICENSE`. No `.desktop` file, as decided.
- [ ] **[Claude]** Check with `tar -tzf` that nothing is outside that directory.

## Phase 5: Local verification

- [ ] **[Claude]** Build the `.deb` and `.rpm` locally, and inspect their file lists with `dpkg -c` and `rpm -qlp`.
- [ ] **[Claude]** Compare the library dependencies with `ldd bin/gamedistapp`. The earlier check showed `libGL`, `libwayland-client` and `libX11`, with the rest coming in through them.
- [ ] **[Claude]** Install the `.deb` in an `ubuntu-22.04` container and in a `debian:12` container. Run `gamedist help` after each install, and check the desktop entry and icon are present.
- [ ] **[Claude]** Install the `.rpm` in a `fedora:<N>` container with `dnf install --nogpgcheck`, then run `gamedist help`.
- [ ] **[Claude]** Remove each package and confirm that every file it installed is gone, while the runtime libraries stay.
- [ ] **[Claude]** Confirm the desktop database and icon cache refreshes do nothing harmful when their tools are missing.
- [ ] **[You]** Look at the desktop entry on a real Linux desktop: it should appear under Utilities after the install.

## Phase 6: Release workflow (`.github/workflows/fsl-release.yml`)

- [ ] **[Claude]** Add a step that downloads the pinned `nfpm` and checks its SHA-256 checksum. The step fails if the checksum does not match.
- [ ] **[Claude]** Add the packaging step after `make build`. It builds the packages from `bin/`, and it is skipped for tags with a hyphen.
- [ ] **[Claude]** Add a `.deb` check job on `ubuntu-22.04` and a second one in a `debian:12` container. Each installs the package and runs `gamedist help`.
- [ ] **[Claude]** Add an `.rpm` check job in a pinned `fedora:<N>` container, using `dnf install --nogpgcheck`.
- [ ] **[Claude]** Add `desktop-file-validate` to the checks, and confirm the installed entry exists.
- [ ] **[Claude]** Make the release creation depend on the checks, so a failed check means no draft release.
- [ ] **[Claude]** Attach the `.deb`, `.rpm` and `.tar.gz` with the zip, using the existing `softprops/action-gh-release` step.
- [ ] **[Claude]** Check that the release notes step and the draft release behaviour are unchanged.

## Phase 7: Testing the workflow

- [ ] **[You]** Run the workflow as chosen in Phase 0, and tell me the result.
- [ ] **[Claude]** Read the run log and fix anything that fails. Repeat until the checks pass.
- [ ] **[You]** Delete any test release and tag afterwards, as `docs/releasing.md` describes. Tell me when that is done.

## Phase 8: Documentation

- [ ] **[Claude]** `docs/building.md`: add the Linux packages section, the packaging tools with their install commands, the pinned `nfpm` version, and the `make package-linux` target.
- [ ] **[Claude]** `docs/releasing.md`: list the new artifacts in the workflow steps, name the pinned Fedora release and say how to bump it, and describe the hyphen-tag rule.
- [ ] **[Claude]** `README.md`: add the `.deb` and `.rpm` install lines, the `.rpm` `--nogpgcheck` note, and "Fedora N and later" with the pinned N.
- [ ] **[Claude]** Check that the README links still work on GitHub once the files are on `master`.

## Phase 9: Review and release

- [ ] **[Claude]** Run `make lint` and `go vet` with the `desktop` tag. Run `go test ./src/...` as `CONTRIBUTING.md` describes. Do not run `make test`.
- [ ] **[Claude]** Check the requirements file against the result, and mark each requirement that is met.
- [ ] **[You]** Review the diff and decide on the commit. Commit only when you ask.
- [ ] **[You]** Push the branch and open the pull request. Merge and tag the release when you are ready.
- [ ] **[You]** After the release, check the attached packages on a real Ubuntu, Fedora and Debian machine if you have one.
