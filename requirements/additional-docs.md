# Additional Documentation Requirements

## Overview

`README.md` is the single end-user guide for the CLI and the desktop app. It now covers installation, usage, the desktop
app and CSV output, but it is long and mixes audiences. Some material is missing or only lives in the code, the
requirements files or the local `CLAUDE.md`. This plan adds focused documents so that each audience can find what it needs
without reading the whole README.

## Scope

- Documentation only. No change to the CLI or desktop app behaviour.
- `README.md` stays the entry point and links to the new documents. It is still bundled into `dist/gamedist.zip`, so the
  new documents must be useful without the rest of the repository, or be linked by URL.
- `CLAUDE.md` stays local and git-ignored. The project facts it held have been moved into the new documents, and it now links to them so that future Claude Code sessions can find them.

## Candidate Documents

| Document | Audience | Purpose |
|---|---|---|
| `docs/desktop-app.md` | End users | Full desktop app guide, with screenshots of the inputs, grid, summary and menu |
| `docs/algorithm.md` | Users and contributors | How a distribution is built: fixtures, `MustPlay`, spacing, max players together, attempts and restarts |
| `docs/building.md` | Contributors | Toolchain, Linux and Windows cross-build prerequisites, `make` targets, Wine testing and the `.syso` icon step |
| `docs/releasing.md` | Maintainers | Release workflow, the `VERSION` file, draft releases, test tags, cleaning up test releases, and where release notes are written (GitHub releases) |
| `CONTRIBUTING.md` | Contributors | Branch and pull request flow, licence header, lint and test commands, and what not to run (`make test`) |

These are a starting set. Any of them can be dropped or merged during the work. There is no changelog file: GitHub release
notes serve that purpose.

## Status

Done (not yet committed):

- `CONTRIBUTING.md` (root): licence header rule, Go version and toolchain, build, run, test and lint commands, the `make test` warning, the version source, Fyne development notes, release notes location, and the Mac builder TODO.
- `docs/building.md`: toolchain, Linux and Windows build packages, `make` targets and outputs, the Windows `.syso` icon step, and the Wine run command with its caveat.
- `docs/algorithm.md`: how a distribution is built, including restarts, attempts, the maximum players together minimum with the worked example, and fixture spacing.
- `CLAUDE.md` (local, git-ignored): the moved facts were removed and replaced by a "Project documentation" section that links to each new file and describes its contents, plus the points that were not moved (CLI arguments, fragile state, distributor code map, desktop app architecture).
- The Mac builder TODO is in both `CONTRIBUTING.md` and `README.md`.

Not yet done:

- `docs/releasing.md`. It is linked from `CONTRIBUTING.md` and `CLAUDE.md`, so those links are broken until it exists.
- `docs/desktop-app.md`, and its screenshots in `docs/images/`.
- Links to the new documents from `README.md`.
- Checking that the commands and file names in the new documents match the repository.

## Functional Requirements

- Each document states its audience and the version it describes in its first lines.
- The desktop app guide includes screenshots of the inputs, the grid, the summary and the menu. They are PNG files in `docs/images/`, taken on Linux. Each caption names the platform, and the guide notes that the Windows appearance has not been checked on a real Windows machine.
- The desktop app guide and the algorithm note cover the same wording and limits as the code (`src/desktop/model`,
  `src/distributor`), not a copy of an older version.
- The algorithm note explains the worked example (24 fixtures, 7 players, 3 per fixture: minimum 4, but spacing needs 5)
  so that users understand why the cap can be raised.
- The building guide lists the exact packages used by the release workflow and the Makefile, including the Ubuntu 22.04
  runtime libraries (`libgl1`, `libwayland-client0`, `libx11-6`).
- The releasing guide matches `.github/workflows/fsl-release.yml` and the Makefile. It must not describe steps that the
  workflow does not run.
- The releasing guide says that release notes are written on the GitHub release, and links to the releases page.
- `README.md` links to each new document from the section it relates to, and no content is duplicated in both places
  beyond a one-line summary.
- Links between documents are relative, and each one works on GitHub.

## Non-Functional Requirements

- Plain Markdown, no generated site, no new tooling.
- Keep each document short enough to read in one sitting. Prefer a link to the code over a copy of the code.
- Use British spelling, as in the existing README and requirements.

## Open Questions

- Should `docs/` be created, or should the new files sit in the repository root beside `README.md`? **Answered:** create `docs/`. Files other than `README.md` go there. `CONTRIBUTING.md` stays in the root, as GitHub expects it there.
- Is a changelog wanted, or do GitHub release notes already serve that purpose? **Answered:** GitHub release notes serve that purpose, so there is no `CHANGELOG.md`. The developer documentation (`docs/releasing.md`) says where release notes are written.
- Are screenshots wanted in the desktop app guide? **Answered:** yes. They are taken on Linux and the caption says so, because the Windows look has only been checked under Wine.
- Should the Mac builder TODO in `README.md` be moved into `CONTRIBUTING.md`? **Answered:** copy it to `CONTRIBUTING.md`, and keep it in `README.md`. `CONTRIBUTING.md` is created with a TODO section holding that item.

## Acceptance

- Each document in the agreed set exists, is linked from `README.md`, and renders on GitHub.
- A reader following `docs/building.md` on Ubuntu 22.04 can build `bin/gamedist` and `bin/gamedistapp`. Other distributions and versions are not tested. The original author is successfully able to build on Ubuntu 26.04.1 LTS.
- The commands and file names in the new documents exist in the repository.
