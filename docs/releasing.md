# Releasing

> **For the package maintainer only.** This guide is for the person who publishes releases. It is not needed to build or use the app.

Audience: maintainers who publish a new version. This guide describes the release process as set up in
`.github/workflows/fsl-release.yml` and the `Makefile`.

## Versions

The version is the root `VERSION` file, for example `v1.1.0`. `make build` reads it and injects it into both the CLI and
the desktop app with `-ldflags "-X main.version=..."`, so the binaries report the value in `VERSION`.

Before you tag a release, make sure the tag name matches `VERSION`. The workflow names the archive and the release after the
tag, but the binaries report `VERSION`, so the two must agree.

The Linux package names use the version without the leading `v`. For `v1.1.0` the packages are `gamedist_1.1.0-1_amd64.deb`,
`gamedist-1.1.0-1.x86_64.rpm` and `gamedist-1.1.0-linux-amd64.tar.gz`. The zip keeps the `v`, as `gamedist-v1.1.0.zip`.

## Release steps

1. Make sure the code you want to release is on `master`, and that `make build` runs on your machine.
2. Set `VERSION` to the new version, and commit it.
3. Create a tag with the same name, and push it:

   ```
   git tag v1.1.0
   git push origin v1.1.0
   ```

4. Wait for the **Draft FSL Release with Build Artifacts** workflow to finish. You can follow it on the Actions tab.
5. Open the draft release on the Releases page, check the notes and the attached files, then publish it. You can also
   publish from the command line:

   ```
   gh release edit v1.1.0 --draft=false
   ```

The workflow runs only when a tag starting with `v` is pushed. Pushing a branch does not start it. A manual run is
described under [Test releases](#test-releases).

## What the workflow does

The workflow has five jobs. Each one runs only when the previous jobs have passed.

- **build** runs on `ubuntu-22.04`. That image has glibc 2.35, so the Linux desktop app runs on Ubuntu 22.04 and later, and on
  Debian 12 and later. It:
  - installs the desktop app build packages: `libgl1-mesa-dev`, `xorg-dev`, `libwayland-dev`, `libxkbcommon-dev` and
    `gcc-mingw-w64-x86-64`
  - runs `make build`, which produces the four binaries and `dist/gamedist.zip` (see [building.md](building.md))
  - copies the zip to `gamedist-<tag>.zip`. That zip already holds the binaries and `README.md`, so it is not packed again
  - downloads the pinned `nfpm` and checks its SHA-256 checksum, then runs `make package-linux` to build the `.deb`, the `.rpm`
    and the `.tar.gz`
- **check-deb-ubuntu**, **check-deb-debian** and **check-rpm-fedora** each install one package in a clean container:
  `ubuntu:22.04`, `debian:12`, and `fedora:<N>`. Each runs `packaging/scripts/check-install.sh`, which runs `gamedist help`,
  validates the desktop entry, checks the icons, and fails if a library is missing.
- **draft-release** creates a **draft** GitHub release for the tag. It attaches the zip, the `.deb`, the `.rpm` and the
  `.tar.gz`, and generates the release notes from the commits and pull requests since the previous release. It runs only for
  tags, and only when the checks have passed.

A tag with a hyphen, such as `v0.0.1-test`, builds the zip only. Its package checks are skipped, and it still gets a draft
release.

The workflow does not build for macOS, and it does not publish the release. Publishing is a manual step.

## Pinned versions

The workflow pins three values in its `env` block. Change them by hand:

- `NFPM_VERSION` and `NFPM_SHA256`: the `nfpm` release and the SHA-256 of its `nfpm_<version>_Linux_x86_64.tar.gz` file, from
  the release's `checksums.txt`.
- `FEDORA_VERSION`: the Fedora release that the `.rpm` check runs in. Bump it when a newer Fedora becomes the current stable
  release. Check the package on the new release before you tag.

## Release notes

Release notes are written on the GitHub release. The workflow generates the "What's Changed" section from the commits and
pull requests. Edit the draft to add anything the generated list does not cover. There is no changelog file in the
repository.

## Test releases

To check a change to the workflow, run it by hand from the branch. A manual run builds the packages and runs all the checks,
but creates no tag and no release:

```
gh workflow run fsl-release.yml --ref <branch>
```

To test the release itself, push a test tag. A tag with a hyphen, such as `v0.0.1-test`, builds the zip only, so use a tag
without a hyphen if you need to test the packages. The workflow creates a draft release for it. Check the files, then delete
the test release and its tags:

```
gh release delete v0.0.1-test --yes
git push origin --delete v0.0.1-test
git tag --delete v0.0.1-test
```

Delete the test tag and release as soon as you have checked them. A test tag is a real tag on the remote, and a draft release
is visible to anyone with write access to the repository.
