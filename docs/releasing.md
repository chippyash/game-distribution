# Releasing

> **For the package maintainer only.** This guide is for the person who publishes releases. It is not needed to build or use the app.

Audience: maintainers who publish a new version. This guide describes the release process as set up in
`.github/workflows/fsl-release.yml` and the `Makefile`.

## Versions

The version is the root `VERSION` file, for example `v1.1.0`. `make build` reads it and injects it into both the CLI and
the desktop app with `-ldflags "-X main.version=..."`, so the binaries report the value in `VERSION`.

Before you tag a release, make sure the tag name matches `VERSION`. The workflow names the archive and the release after the
tag, but the binaries report `VERSION`, so the two must agree.

## Release steps

1. Make sure the code you want to release is on `master`, and that `make build` runs on your machine.
2. Set `VERSION` to the new version, and commit it.
3. Create a tag with the same name, and push it:

   ```
   git tag v1.1.0
   git push origin v1.1.0
   ```

4. Wait for the **Draft FSL Release with Build Artifacts** workflow to finish. You can follow it on the Actions tab.
5. Open the draft release on the Releases page, check the notes and the attached archive, then publish it. You can also
   publish from the command line:

   ```
   gh release edit v1.1.0 --draft=false
   ```

The workflow runs only when a tag starting with `v` is pushed. Pushing a branch does not start it.

## What the workflow does

- Runs on `ubuntu-22.04`. That image has glibc 2.35, so the Linux desktop app runs on Ubuntu 22.04 and later, and on
  Debian 12 and later.
- Installs the desktop app build packages: `libgl1-mesa-dev`, `xorg-dev`, `libwayland-dev`, `libxkbcommon-dev` and
  `gcc-mingw-w64-x86-64`.
- Runs `make build`, which produces the four binaries and `dist/gamedist.zip` (see [building.md](building.md)).
- Packs the contents of `dist/` into `dist-<tag>.tar.gz`.
- Creates a **draft** GitHub release for the tag. It attaches the archive, and generates the release notes from the commits
  and pull requests since the previous release.
- Adds a licence notice to the release body. It gives the release date, the licence (FSL-1.1), and the date on which that
  tag converts to Apache License 2.0, two years after release.

The workflow does not build for macOS, and it does not publish the release. Publishing is a manual step.

## Release notes

Release notes are written on the GitHub release. The workflow generates the "What's Changed" section from the commits and
pull requests. Edit the draft to add anything the generated list does not cover. There is no changelog file in the
repository.

## Test releases

To check a change to the workflow, push a test tag such as `v0.0.1-test`. The workflow creates a draft release for it. Check
the archive, then delete the test release and its tags:

```
gh release delete v0.0.1-test --yes
git push origin --delete v0.0.1-test
git tag --delete v0.0.1-test
```

Delete the test tag and release as soon as you have checked them. A test tag is a real tag on the remote, and a draft release
is visible to anyone with write access to the repository.
