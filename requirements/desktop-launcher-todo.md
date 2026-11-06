# Desktop Launcher Implementation To-Do

This is the implementation plan for `requirements/desktop-launcher.md`. Each task is marked with who carries it out:

- **[Claude]**: I can do this in the repository.
- **[You]**: needs an action from you, such as testing on a desktop, or deciding on a commit.
- **[Both]**: I do the work, and you check or decide something.

Status: not started. Work is on the branch `feature/desktop_launcher`, created from `master`.

## Decisions this plan relies on

These come from the answers in `requirements/desktop-launcher.md`:

- The question is asked after the splash, on every start, until the user chooses "Create" or "Never ask".
- "Create" and "Never ask" store the choice under the Fyne preference key `desktopShortcut`, with the values `created` and `never`.
- The shortcut and menu entry are written by the app, with the `X-Gamedist-Managed=true` marker, `Exec` set to the quoted real path of the binary, and `Icon` set to an absolute path.
- The app rewrites its own files when any key differs, and it never touches a file without the marker.
- The app refuses to create a shortcut from the system temporary folder, and leaves existing ones that point there alone.
- The success message includes one sentence about desktops that hide Desktop icons.
- The dialog and menu text are English only.

## Phase 0: Prerequisites

- [x] **[Claude]** Check the Fyne v2.8.2 preferences and dialog APIs that the question needs: the modal dialog with three buttons, and whether dismissing it can be detected. Done, in Fyne v2.8.2. The modal dialog comes from `NewCustomWithoutButtons` with `SetButtons`, which takes any number of buttons. `SetOnClosed` runs on every close, so each button sets its answer before it hides the dialog, and the close handler treats no answer as "Not now". Preferences have `String` and `SetString`, as the plan assumes.
- [x] **[Claude]** Check which tools are present in the test containers: `xdg-user-dir` (`xdg-user-dirs`), `gio` (`glib2`), and `update-desktop-database` (`desktop-file-utils`). Done. None of `xdg-user-dir`, `gio` or `update-desktop-database` is present in `ubuntu:22.04` or `debian:12`. In `fedora:45`, only `gio` is present. The test images need `xdg-user-dirs`, `libglib2.0-bin` (Ubuntu and Debian) and `desktop-file-utils`. I'll confirm those package names when the container tests are written in Phase 5.
- [ ] **[You]** Have a GNOME session available for the trusted-mark check, and a KDE session if you have one. Tell me which you can use. Checked: KDE Plasma 6 on Wayland, on Ubuntu 26.04.1 LTS. GNOME is not installed, so the GNOME trusted-mark check is not possible here. Tell me if you want GNOME installed for it.
- [x] **[Claude]** Research the GNOME trusted mark. The sources are mixed: `gio set ... metadata::trusted true` is reported to work on some GNOME setups, and to fail with "not supported" on others. The manual Permissions step is the fallback. See [dropdesk](https://github.com/Enumber/dropdesk), [the GNOME Discourse thread](https://discourse.gnome.org/t/ubuntu-20-04-lts-gnome-desktop-files/3878) and [Launchpad bug 572918](https://bugs.launchpad.net/ubuntu/+source/meta-gnome2/+bug/572918/+index). Not verified here, because GNOME is not installed.

## Phase 1: Model logic (`src/desktop/model`, no Fyne import)

- [x] **[Claude]** Desktop folder resolution: `xdg-user-dir DESKTOP`, then `~/Desktop` if it exists, with a result equal to `$HOME` treated as no folder. Injectable command runner and home folder, so the tests need no real desktop. Done: `DesktopFolder` in `launcher.go`, with an injected command runner and home folder.
- [x] **[Claude]** Desktop entry rendering: the keys from the requirements, `Exec` quoted and escaped for paths with spaces, `Icon` as an absolute path, and the `X-Gamedist-Managed=true` marker. Done: `Shortcut.Render`, with `Exec` quoted and escaped, and tests for spaces, quotes, dollar signs and backslashes.
- [x] **[Claude]** Marker parsing and the refresh comparison: a file with the marker is rewritten when any key differs, and a file without it is never changed. Done: `ParseEntry`, `IsManaged` and `Plan`, which returns create, rewrite, keep or conflict.
- [x] **[Claude]** The stored choice: map `created`, `never`, and no value, to the question's state. Done: `ShouldAsk` and the `ChoiceKey`, `ChoiceCreated` and `ChoiceNever` constants.
- [x] **[Claude]** The temporary-folder rule: a binary path under `os.TempDir()` is refused for creation, and an existing shortcut under it is left alone and reported. Done: `UnderTempDir`, which does not match a sibling folder with the same prefix.
- [x] **[Claude]** Table-driven tests for every rule above, including the `$HOME` case and paths with spaces. Done: `launcher_test.go`. The model tests pass, and golangci-lint passes on the package.

- [x] **[Claude]** Desktop detection from `XDG_CURRENT_DESKTOP`: `IsGNOME(value string) bool`, which reads the colon-separated list. Done: tests cover GNOME, a list that contains GNOME, a GNOME variant, lower case, KDE, XFCE and an empty value. Tests pass, and lint is clean.
## Phase 2: Files and icon (`src/desktop/model`, no Fyne import)

- [x] **[Claude]** Embed the 256 pixel icon in the binary, and write it to `~/.local/share/icons/hicolor/256x256/apps/gamedistapp.png` when the package has not installed it, or when its bytes differ. Done: `src/desktop/launcher_icon.go` embeds `src/desktop/resources/gamedistapp-256.png`, a copy of the packaged icon. `EnsureIcon` writes it when the package has no icon, or when the bytes differ. The embed is unused until Phase 3, so golangci-lint reports it as unused until then.
- [x] **[Claude]** Write the Desktop shortcut and the user menu entry with mode `0755`, and the icon with `0644`. Done: `WriteEntry` writes with mode `0755`, and `EnsureIcon` writes the icon with `0644`. A conflicting file is left alone.
- [x] **[Claude]** Delete the app's own user menu entry when a system entry exists at `/usr/share/applications/gamedistapp.desktop`. Done: `RemoveOwnMenuEntry`, which deletes only a file with the marker, and only when the system entry exists.
- [x] **[Claude]** Run `update-desktop-database` on `~/.local/share/applications/` after a change, when the tool is present. Done: `RefreshDatabase`, which does nothing when the tool is absent.
- [x] **[Claude]** Mark the Desktop shortcut trusted with `gio set ... metadata::trusted true` after creating or rewriting it, on every desktop, when `gio` is present. If `gio` fails, return the manual Permissions step for the success message or the menu item. Done: `MarkTrusted`, which does nothing when `gio` is absent, and returns the error when `gio` fails. Phase 3 calls it after each create or rewrite.
- [x] **[Claude]** Tests using temporary folders: create, refresh, conflict with a file without the marker, read-only folder, and the menu entry added and removed. Done: `launcher_files_test.go`. The model tests pass, and lint is clean on the model package.

## Phase 3: Start-up and dialogs (`src/desktop`)

- [x] **[Claude]** After the splash, run the refresh, then ask the first-run question unless the stored choice is `created` or `never`, the app runs as root, or there is no Desktop folder. Done: `startLauncher` runs from `splashDone`. It refreshes the app's own entries first, then asks the question if `ShouldAsk` allows it. It does nothing as root or on Windows.
- [x] **[Claude]** The first-run dialog: modal, three buttons "Create", "Not now" and "Never ask". Dismissing it counts as "Not now". Done: `askLauncher`, a modal dialog with the buttons Create, Not now and Never ask. Dismissing it counts as Not now.
- [x] **[Claude]** The create action: the "Also add to the application menu" checkbox, shown only when no system entry exists and only when a Desktop folder exists. Report the shortcut and the menu entry separately. Store `created` only when the Desktop shortcut is created. Done: `createLauncher` and `showLaunchReport`. The first-run dialog creates the Desktop shortcut only. The menu item offers the menu entry checkbox when no system entry exists. The stored value is `created` when the Desktop shortcut is created or already current. A conflicting file does not store it.
- [x] **[Claude]** The success message with the sentence about hidden Desktop icons, shown on GNOME only. Done: `showLaunchReport` adds the sentence on GNOME only, found with `IsGNOME`.
- [x] **[Claude]** The "Create desktop shortcut" menu item, which says so in one line when everything is up to date, and stores `created` when it creates the Desktop shortcut. Done: `onCreateShortcut` is in the menu on Linux. It says so in one line when everything is up to date, and it stores `created` when it creates the Desktop shortcut.
- [x] **[Claude]** Hide the menu item on Windows, and never ask on Windows. Done: the menu item is added only when the launcher is enabled, which is false on Windows. The question is never asked there.
- [x] **[Claude]** Refresh failures on start go to the log and the menu item, with no dialog. Done: `refreshLauncher` logs failures and keeps them for the menu item, with no dialog on start. The menu item shows the last problem in its dialog.

## Phase 4: Tests

- [x] **[Claude]** UI tests with `test.NewApp()`: the first-run question appears only in the right states, the stored choice is written for each button, and the splash order is respected. Done: `launcher_ui_test.go`, using `test.NewApp()`. The tests cover: the question appears only after the splash; it is asked only in the right states (no choice, root, no Desktop folder, Windows, and the stored choices); the refresh runs before it; each answer stores the right value; a foreign file or a temporary copy stores nothing; the menu entry is added only when no system entry exists; gio and the database tool are called; the menu item is offered only when enabled; and the report text. The buttons' wiring is tested through `applyLauncherAnswer`, and the buttons were also checked by hand in Phase 5. The tests use temporary folders and a disabled-or-fake environment, so they never touch the real Desktop.
- [x] **[Claude]** Run `go test -tags desktop ./src/desktop/...`, and `golangci-lint run --build-tags desktop ./src/desktop/...`. Done: the desktop and model tests pass, and lint is clean. `go test ./src/...` passes, and the Windows build vets clean.

## Phase 5: Manual testing

- [x] **[Claude]** Build the app and run it in a container with a virtual display, checking that the question appears after the splash and that each button stores the right value. Done in a Fedora 45 container with a virtual display, as a non-root user. The question appears after the splash, as a modal dialog. "Never ask" stores `never`, and the question does not come back. "Not now" stores nothing, and the question returns on the next start. "Create" makes the Desktop shortcut and stores `created`. Without GVFS, `gio` reports the attribute is not supported, so the success message gives the manual Permissions step. The dismiss path (closing the dialog without a button) is not tested.
- [ ] **[Claude]** Check the `.deb` install, then the `.tar.gz` install, the move of the `.tar.gz` folder, and the upgrade, and confirm the shortcut is refreshed in each case. Partly done: the `.tar.gz` and `.rpm` were built from the current code and installed in the container. The `.tar.gz` shortcut was refreshed to its folder. Installing the `.rpm` moved `Exec` to `/usr/bin/gamedistapp` and removed the app's own user menu entry. Removing the `.rpm` left the shortcut in place, pointing at the missing file. A move of the app was also checked. The `.deb` was not installed: it needs an Ubuntu 22.04 build, which was not made for this test. Also covered: a foreign file in the way is left alone, a copy started from a temporary folder does not change the shortcut, and a conflicting create stores nothing.
- [ ] **[You]** On a real GNOME desktop: answer "Create", double-click the shortcut, and confirm it launches. Then check the hidden-icons message. GNOME is not installed here, and the answer to Round 7 question 4 is no. This check needs a GNOME machine.
- [x] **[You]** On KDE if you have it, the same check. Done for the launch step: on KDE Plasma 6 (Wayland), a test launcher with mode `0755`, `Exec` set to `/usr/bin/gamedistapp` and no trusted mark opened the app on a double-click, with no prompt. The hidden-icons message and the menu entry are still to check.

## Phase 6: Documentation

- [ ] **[Claude]** `docs/desktop-app.md`: a section on the desktop shortcut, covering the first-run question and its three choices, the menu item, the stored choice, and the effect of removing the package.
- [ ] **[Claude]** `docs/releasing.md` and `README.md`: only if the install or removal steps change.

## Phase 7: Review and commit

- [ ] **[Claude]** Check the requirements against the result, and mark each one met, met in the app, or not met.
- [ ] **[You]** Review the diff, and tell me when to commit. Commit only when you ask.
- [ ] **[You]** Push the branch and open the pull request when you are ready.
