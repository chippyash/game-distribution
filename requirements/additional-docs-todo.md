# Additional Documentation Implementation To-Do

This is the implementation plan for `requirements/additional-docs.md`. Each task is marked with who carries it out:

- **[Claude]**: I can do this in the repository.
- **[You]**: needs an action from you, such as a decision, a screenshot on a machine you use, or a commit or push.
- **[Both]**: I do the work, and you check or decide something.

Status: Phases 0 to 3 are complete in the working tree, and nothing is committed yet. Phases 4 to 6 are not started.

## Environment check

- Branch: `feature/additional_docs`, created from `master`.
- `CLAUDE.md` is git-ignored and stays local.
- Linux desktop display is available for screenshots (Phase 4).

## Phase 0: Decisions

- [x] **[You]** `docs/` is created for the new documents. `CONTRIBUTING.md` stays in the repository root.
- [x] **[You]** There is no `CHANGELOG.md`. GitHub release notes serve that purpose, and the releasing guide says where they are written.
- [x] **[You]** Screenshots are included in the desktop app guide. They are taken on Linux, and the captions say so.
- [x] **[You]** The Mac builder TODO is copied to `CONTRIBUTING.md` and kept in `README.md`.
- [x] **[You]** Keep `CLAUDE.md` local. Its project facts move into the new documents, and it links to them.
- [x] **[You]** Acceptance for the build check applies to Ubuntu 22.04 only. Other releases are not tested.

## Phase 1: Requirements

- [x] **[Claude]** Create `requirements/additional-docs.md` with overview, scope, candidate documents, functional and non-functional requirements, open questions and acceptance.
- [x] **[Claude]** Record each answered open question in `requirements/additional-docs.md`.
- [x] **[Claude]** Update `requirements/additional-docs.md` with a Status section that lists what is done and what is not.
- [x] **[Claude]** Create this plan, `requirements/additional-docs-todo.md`.

## Phase 2: Project documentation split

- [x] **[Claude]** Draft three documents in the scratchpad from the project facts in `CLAUDE.md`: `CONTRIBUTING.md`, `docs/building.md` and `docs/algorithm.md`.
- [x] **[Claude]** Copy the drafts into the repository. `CONTRIBUTING.md` keeps the Mac TODO at its end.
- [x] **[Claude]** Remove the transferred items from `CLAUDE.md`, and keep the items no new file covers (CLI arguments, fragile state, distributor code map, desktop app architecture).
- [x] **[Claude]** Add a "Project documentation" section to `CLAUDE.md` that links to each new file and describes its contents, including `requirements/`.

## Phase 3: Check the new documents

- [x] **[Claude]** Check every command and file name in `CONTRIBUTING.md`, `docs/building.md` and `docs/algorithm.md` against the repository, the Makefile and `.github/workflows/fsl-release.yml`. Done: `docs/building.md` was missing `zip` and a Go install note, now fixed. The `make` targets and outputs match the Makefile.
- [x] **[Claude]** Check the worked example and the minimum-cap wording in `docs/algorithm.md` against `src/distributor/distributor.go` and the code that computes the minimum. Done: the CLI gives cap 4 for the example. With spacing, cap 4 fails 5000 attempts and cap 5 succeeds. The wording was corrected to "pair-appearances".
- [x] **[Claude]** Check that the Fyne notes in `CONTRIBUTING.md` still match `src/desktop`, in particular the `helpTip` and `hideGrace` names. Done: both names exist, and the `ExtendBaseWidget` and X11 hint notes match the code.
- [x] **[Claude]** Check the Ubuntu 22.04 build instructions in `docs/building.md` by running the steps in a clean `ubuntu:22.04` container. Done: `make build` exits 0 and produces the five binaries and `dist/gamedist.zip`.
- [x] **[You]** Confirm the Ubuntu 26.04.1 LTS claim in `requirements/additional-docs.md` (Acceptance) or ask for it to be removed. Confirmed: `make build` ran on the local Ubuntu 26.04.1 LTS machine with Go 1.22.0 and exited 0, and produced the five binaries and `dist/gamedist.zip`.

## Phase 4: Desktop app guide and screenshots

- [x] **[Claude]** Write `docs/desktop-app.md` from `README.md`'s desktop section and `src/desktop/model/help.go`. Include the audience and version in its first lines.
- [x] **[Claude]** Create `docs/images/` and take the screenshots of the inputs, the grid, the summary and the menu on Linux. Done: `desktop-main-window.png`, `desktop-inputs.png`, `desktop-results.png` (grid and summary) and `desktop-menu.png`, captured from the running app on `:0` with `xdotool` driving it. The app ran with a throwaway config folder, so the light theme is shown and the saved dark preference is untouched.
- [x] **[Claude]** Add each screenshot to `docs/desktop-app.md` with a caption that names Linux. The guide's opening section notes that the Windows appearance has not been checked on a real Windows machine.
- [x] **[Claude]** Document the menu entry that Fyne adds itself (Quit), which was found while taking the menu screenshot, and note that 1000 attempts can fail for the 24/7/3 example.
- [x] **[You]** Review the screenshots and the guide, and say if any image should be retaken. Reviewed; the guide has one small change from the author.

## Phase 5: Releasing guide

- [x] **[Claude]** Write `docs/releasing.md` from `.github/workflows/fsl-release.yml`, the Makefile and the `VERSION` file. Include the test tags, draft releases, cleaning up test releases, and that release notes are written on the GitHub release.
- [x] **[Claude]** Check that every step in `docs/releasing.md` is one the workflow actually runs. The manual steps (tagging, publishing, deleting test releases) are marked as manual. The Makefile's old `version` and `push` targets were removed, along with the variables only they used, and the guide no longer mentions them.
- [ ] **[You]** Confirm the release steps you follow match the guide.

## Phase 6: Link and finish

- [ ] **[Claude]** Add links from `README.md` to `CONTRIBUTING.md`, `docs/building.md`, `docs/algorithm.md`, `docs/desktop-app.md` and `docs/releasing.md`, each in the section it relates to. Do not remove the Mac TODO from `README.md`. Also trim the README's desktop section to a short summary with a link to `docs/desktop-app.md`, so the two do not duplicate each other. The README section does not list Quit, which the guide now does.
- [ ] **[Claude]** Check that all relative links in the new documents and in `README.md` resolve to files in the repository.
- [ ] **[Claude]** Run `make lint`, and confirm that `make build` still produces `dist/gamedist.zip` with `README.md`.
- [ ] **[Claude]** Update the Status sections of `requirements/additional-docs.md` and this plan.
- [ ] **[You]** Review the diff, then ask me to commit it on `feature/additional_docs`.
- [ ] **[You]** Ask me to push the branch and open a pull request into `master`, and then merge it on GitHub.
- [ ] **[You]** Decide whether to delete `feature/additional_docs` after the merge.
