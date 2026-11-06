# Desktop Launcher Requirements

## Overview

The Linux packages install Game Distributor in the application menu, under Utilities. Some users also want a launcher on
their Desktop. A package cannot ask this question during install: `apt` and `dnf` are non-interactive, the package scripts
run as root, and a package cannot write into each user's home folder. So the desktop app asks the user itself, when it
starts, and can also create the shortcut later from its menu.

## Scope

- Linux only for the first version. Windows shortcuts use a different mechanism and are a separate decision.
- Applies to every Linux install route: the `.deb`, the `.rpm` and the `.tar.gz`.
- The shortcut is a per-user file. The app creates it as the user who runs it, and never as root.
- No change to the CLI.

## Functional Requirements

### Asking

- When the desktop app starts, the splash is shown first. After the splash, the app first refreshes any existing shortcut and
  menu entry, as described under "Keeping the shortcut up to date". Then it asks whether to create a desktop shortcut. It asks
  only if the user has not already chosen "Create" or "Never ask", and the app is not running as root.
- The dialog is modal, so the main window cannot be used until the user answers it. Dismissing the dialog counts as "Not now".
- The dialog has three buttons:
  - **Create** creates the shortcut, as described below. If the Desktop shortcut is created, it stores the choice in the Fyne
    preferences for `com.chippyash.gamedistributor`, under the key `desktopShortcut` with the value `created`, so the question
    stops. If it cannot be created, nothing is stored, the question returns on the next start, and the dialog says why.
  - **Not now** does nothing, and the question is asked again on the next start. There is no limit on how often it is asked.
  - **Never ask** stops the question. It stores `never` under the same key, `desktopShortcut`. Until the user answers, the key
    has no value.
- The app also has a menu item, "Create desktop shortcut". It works at any time, whatever the stored choice. If the shortcut
  and the menu entry are already up to date, it says so in one line, and does nothing else. When it creates the Desktop
  shortcut, it stores `created` under `desktopShortcut`, as "Create" does, so the first-run question stops.
- The app does not ask when it cannot find a Desktop folder. The menu item reports that there is no Desktop folder, and
  offers neither the shortcut nor the menu entry.

### Creating the shortcut

- The Desktop folder is found with `xdg-user-dir DESKTOP`, so a localised name such as `Bureau` works. If `xdg-user-dir` is
  missing, the app uses `~/Desktop`. A folder found this way must exist, and must not be the home folder itself. A result
  equal to `$HOME` counts as no Desktop folder.
- The shortcut is a `.desktop` file in the Desktop folder, named `gamedistapp.desktop`. The app writes the file itself. It
  uses the same name, comment and category as the package's entry.
- `Exec` is the full path of the running binary (`os.Executable`), quoted and escaped as the desktop entry specification
  requires, so a path with spaces works. `os.Executable` resolves symbolic links, so the shortcut points at the real binary. The shortcut then works for the
  `.tar.gz` install, where the binary is not in `/usr/bin`.
- `Icon` is the absolute path of the icon file, not the icon name. The icon file is
  `/usr/share/icons/hicolor/256x256/apps/gamedistapp.png` if the package installed it. Otherwise it is the icon the app writes,
  as described below.
- The file contains the key `X-Gamedist-Managed=true`. The app uses that key to recognise files it wrote.
- The icon is embedded in the binary. If the package has not installed the icon, the app writes it to
  `~/.local/share/icons/hicolor/256x256/apps/gamedistapp.png`. When the file exists but its bytes differ from the embedded icon,
  the app rewrites it, so a new icon in a later version reaches the user.
- The shortcut and the menu entry have mode `0755`, so they are executable. The icon file has mode `0644`.
- After creating or rewriting the Desktop shortcut, the app runs `gio set <file> metadata::trusted true` on it, on every desktop,
  when `gio` is present. GNOME uses this mark to decide whether a double-click may launch a desktop file. KDE launched the
  shortcut without it, so the step does nothing harmful there.
- If `gio` reports that the attribute is not supported, or the command fails, the app says so in the success message or the menu
  item. It then gives the manual step for GNOME: right-click the shortcut, choose Properties, open the Permissions tab, and tick
  "Allow executing file as program".
- The create action also offers "Also add to the application menu", shown only when no system entry exists. When it is
  chosen, the same file, with the same `Icon` and `Exec`, is written to `~/.local/share/applications/`. The menu item shows
  the option, and the first-run dialog does not. The menu entry is not offered when there is no Desktop folder.
- If the shortcut and the menu entry are written separately, each one is reported on its own. If one of them fails, the one
  that succeeded is kept.
- When the Desktop shortcut is created on GNOME, the success message includes one sentence: the shortcut may not be shown if
  Desktop icons are hidden, for example without the Desktop icons extension.

### Keeping the shortcut up to date

- On each start, if the app's own Desktop shortcut or menu entry exists, and its `Exec` or `Icon` no longer matches the running
  binary or the icon file, the app rewrites it. This covers an upgrade to the `.deb`, a move of the `.tar.gz` folder, and a
  change of icon path.
- If a system menu entry exists at `/usr/share/applications/gamedistapp.desktop`, the app deletes its own user menu entry,
  the one in `~/.local/share/applications/` that carries `X-Gamedist-Managed=true`. This avoids two entries in the menu.
- The app rewrites a file when any key it writes differs from the file on disk. This covers `Name`, `Comment`, `Categories`,
  `Exec` and `Icon`.
- After the app rewrites the Desktop shortcut, it sets the trusted mark again, as described above.
- After the app writes, changes or deletes the user menu entry, it runs `update-desktop-database` on `~/.local/share/applications/`
  when the tool is present. If the tool is absent, it does nothing.
- The app never changes or deletes a file that does not carry `X-Gamedist-Managed=true`. If such a file is in the way, the app
  leaves it alone and reports that it did.

### Errors

- If the shortcut cannot be created, for example because the Desktop folder is read-only, or a file of that name is in the
  way, the app shows the reason in a dialog and carries on. Nothing is stored, so the question returns on the next start.
  It never stops the app from starting.
- If a refresh on start fails, the app writes the reason to its log and shows it in the menu item, and does not show a dialog.
- If the running binary is under the system temporary folder (`os.TempDir()`), the app does not create a shortcut. The create
  action shows a dialog that says why, and nothing is stored.
- If an existing shortcut or menu entry has an `Exec` under the system temporary folder, the refresh leaves it alone and reports
  it in the menu item. It is not rewritten to the temporary path.
- The app never runs as root to create the shortcut. If it is running as root, it does not ask the question at all.

### Windows

- On Windows the question is not asked, and the "Create desktop shortcut" menu item is hidden.

### Removal

- Removing the package does not remove a shortcut the user created. The package does not know about it. After
  `apt remove` or `dnf remove`, the shortcut stops working, and the documentation says so.
- `docs/desktop-app.md` has a section on the desktop shortcut. It covers the first-run question and its three choices, the
  menu item, the stored choice, and the effect of removing the package.

## Non-Functional Requirements

- The question must not get in the way. It appears after the splash, and it is answered or dismissed in one step.
- The logic for paths, file contents, the stored choice and the refresh rule is kept in `src/desktop/model`, where it does
  not need Fyne, so it is tested without a display. The dialogs and the menu item are in `src/desktop`.
- No new runtime dependency for the desktop app. `xdg-user-dir` and `gio` are used when present, and skipped when absent.
- The trusted step is checked by hand on GNOME, because it cannot be tested without a GNOME session. KDE is checked by hand too.
- The desktop is found from the `XDG_CURRENT_DESKTOP` environment variable, which can hold several names separated by colons. The
  app uses it only for the GNOME-specific wording. It never uses it to decide whether to run `gio`.
- The dialog and menu text are in English only in this version. Translation is a later decision.
- Use British spelling, as in the existing documentation.

## Open Questions

The first set of questions is answered. The answer to the first one, "Should the question appear on the first run only?", is
replaced by review answer 1 below.

- Should the question appear on the first run only, or on every run until the user answers? **Answered:** first run only,
  with "Do not ask again" stored. *Replaced by review answer 1.*
- If the user answers "Do not ask again" and later removes the shortcut, should the app offer it again? **Answered:** no.
  The menu item is the way to recreate it.
- Should the shortcut also appear in the menu for the `.tar.gz` user? **Answered:** yes, if the user agrees, as described above.
- Should the app ask on Windows too? **Answered:** not in this version.
- Should the question be skipped when there is no Desktop folder? **Answered:** yes, and the menu item reports that there is no
  Desktop folder.

## Review Questions

### Round 1

1. **"Not now" versus "Do not ask again".** **Answered:** three buttons, "Create", "Not now" and "Never ask". "Not now" asks
   again on the next launch.
2. **Duplicate menu entries.** **Answered:** the user menu entry is written only when no system entry exists.
3. **Stale shortcut after an upgrade or a move.** **Answered:** the app refreshes a shortcut it wrote itself, and leaves any
   other file alone.
4. **An existing file with the same name.** **Answered:** the same rule as question 3.
5. **Windows.** **Answered:** the menu item is hidden on Windows for now.
6. **No Desktop folder.** **Answered:** skip and report.
7. **Dialog behaviour.** **Answered:** modal. Dismissing it counts as "Not now".
8. **Paths with spaces in `Exec`.** **Answered:** yes, `Exec` is quoted.
9. **Shortcut after package removal.** **Answered:** the package does not remove it, and it stops working.

### Round 2

1. **Where "Create" is stored.** **Answered:** "Create" stores the choice in the same preference as "Never ask".
2. **`xdg-user-dir` returns the home folder.** **Answered:** a result equal to `$HOME` counts as no Desktop folder.
3. **Icon path in `Icon=`.** **Answered:** the absolute path of the icon file.
4. **Refresh of the menu entry.** **Answered:** yes, the menu entry is refreshed in the same way as the Desktop shortcut.
5. **Menu entry without a Desktop folder.** **Answered:** no. Without a Desktop folder, neither the shortcut nor the menu entry is
   offered.
6. **Partial failure.** **Answered:** yes, each one is reported separately, and the one that succeeded is kept.
7. **"Not now" every start.** **Answered:** keep it as it is, with no limit.

## Review Questions, Round 3

These came from reviewing the requirements again. Each has a recommendation. Please answer each one.

1. **What is stored when "Create" fails.** The requirements store the choice when "Create" is chosen, but do not say what
   happens if the shortcut cannot be written, or the Desktop file already exists and is not the app's. Should nothing be
   stored in that case, so the question returns on the next start? **Answered:** yes, but the dialog should say what went
   wrong, so the user knows why it came back.
2. **A stale user menu entry after a `.deb` install.** A `.tar.gz` user may already have a menu entry in
   `~/.local/share/applications`. After a `.deb` install the system entry exists too, so the menu shows two Game Distributor
   entries. Should the app delete its own user menu entry once a system entry exists? **Answered:** yes. The rule applies
   only to a file that carries `X-Gamedist-Managed=true`.
3. **The menu item when the shortcut is already current.** If the user chooses "Create desktop shortcut" and the shortcut and
   menu entry are already up to date, should the app say so, or simply do nothing? **Answered:** say so, in one line.
4. **Stored preference name.** The requirements do not name the preference key. Should it be `desktopShortcut`, with the values
   `created` and `never`, and no value while the user has not answered? **Answered:** yes.
5. **Which package icon counts as "installed".** The requirements say to use the package's icon if the package installed one,
   but do not give the path. Should the app check `/usr/share/icons/hicolor/256x256/apps/gamedistapp.png`, and use the icon it
   writes under `~/.local/share/icons` only if that file is missing? **Answered:** yes.
6. **Symbolic links.** `os.Executable` returns the resolved path, so if the user starts the app through a symbolic link, the
   shortcut uses the real file. Is that acceptable? **Answered:** yes. The shortcut then follows the real binary, and the
   refresh rule keeps it in step.

## Review Questions, Round 4

These came from reviewing the requirements again. Each has a recommendation. Please answer each one.

1. **A shortcut made from the menu item.** If the user chooses "Create desktop shortcut" from the menu, the requirements do not
   say whether that stores `created`. Without it, the first-run question keeps asking after the user has already made the
   shortcut. Should the menu item store `created` when the Desktop shortcut is made? **Answered:** yes.
2. **Refresh failures on start.** The refresh runs on every start. If it fails, for example because the Desktop folder is read-only,
   should the app show a dialog each time, or log the failure and show it only in the menu item? **Answered:** log it, and show
   it in the menu item, so the user is not nagged on every start.
3. **What the refresh compares.** The refresh rewrites a file when `Exec` or `Icon` changes. Should it also rewrite when `Name`,
   `Comment` or `Categories` change, so that a new version's text reaches the shortcut? **Answered:** yes. The app rewrites
   the file when any key it writes differs from the file on disk.
4. **File permissions.** The requirements say the file is marked executable, but not the mode. Should the mode be `0755` for the
   shortcut and the menu entry, and `0644` for the icon? **Answered:** yes.
5. **A temporary or mounted binary.** If the app is run from a temporary folder, or from a mounted package image that goes away
   later, the shortcut would point at a path that disappears. Should the app refuse to create the shortcut when the binary is under
   the system temporary folder, and report why? **Answered:** yes.
6. **Dialog wording and language.** Are the dialog and menu text in English only for this version, as the rest of the app is?
   **Answered:** yes. Translation is a later decision.

## Review Questions, Round 5

These came from reviewing the requirements again. Each has a recommendation. Please answer each one.

1. **GNOME trust after a rewrite.** The refresh rewrites the Desktop shortcut when it changes. The trusted mark is set once at
   creation, and it is not clear that it survives a rewrite. Should the app set the trusted mark again after every rewrite?
   **Answered:** yes.
2. **Menu database after writing the user menu entry.** GNOME and KDE may not show a new or changed entry in
   `~/.local/share/applications` until the desktop database is updated. Should the app run `update-desktop-database` on that
   folder after it writes or deletes the user menu entry, when the tool is present? **Answered:** yes. This is the same
   approach the package scripts use.
3. **Icon updates.** The embedded icon can change between versions, but the requirements only write the icon when it is
   missing. Should the app rewrite `~/.local/share/icons/.../gamedistapp.png` when its bytes differ from the embedded icon?
   **Answered:** yes, the same rule as the other keys.
4. **Documentation.** The requirements say "the documentation says so" about removal, but do not say which documents change.
   Should `docs/desktop-app.md` gain a section on the desktop shortcut, covering the first-run question, the menu item, the
   stored choice, and the effect of removing the package? **Answered:** yes, as part of the implementation.
5. **A shortcut that points at a temporary path.** If an existing shortcut's `Exec` points under the system temporary folder,
   for example after the app was run from a temporary copy, should the refresh leave it alone and report it, rather than
   rewrite it to the temporary path? **Answered:** yes.

## Review Questions, Round 6

The requirements have covered most of the behaviour now. Most of what is left is wording. These three questions need your
answer before the implementation plan is written.

1. **Desktops that hide Desktop icons.** Recent GNOME releases do not show files in the Desktop folder unless an extension is
   enabled. The shortcut is created, but the user may not see it and think it failed. Should the success message say that the
   shortcut may not be shown on desktops that hide Desktop icons? **Answered:** yes, one sentence.
2. **Order at start.** The refresh runs on each start, and the question appears after the splash. Should the refresh run first,
   so the question sees the current state? **Answered:** after the splash screen has shown. The refresh runs after the splash, and the question follows it.
3. **Next step.** Should I now write the implementation plan, `requirements/desktop-launcher-todo.md`, and stop the review rounds
   unless you find a problem? **Answered:** yes.

## Review Questions, Round 7

Testing on KDE showed that the trusted mark is not needed there. GNOME may need it, and the evidence is mixed. The
requirements now run `gio set ... metadata::trusted true` on every desktop, and detect GNOME only for wording. These questions
confirm that.

1. **Run `gio` on every desktop.** Should the app run the trust step on every desktop, rather than only on GNOME? **Answered:**
   yes. This avoids desktop detection for the trust step, and KDE is unaffected.
2. **Fallback when `gio` fails.** If `gio` cannot set the mark, should the app give the manual Permissions step in the success
   message or the menu item? **Answered:** yes.
3. **GNOME-only wording.** Should the hidden-icons sentence appear only on GNOME, found from `XDG_CURRENT_DESKTOP`? **Answered:**
   yes. This is the only place the desktop is detected, so it is needed for that sentence alone.
4. **A GNOME session to test on.** GNOME is not installed on this machine. Should I install GNOME alongside KDE for the trust check,
   or will you test on a GNOME machine? **Answered:** no. GNOME is not installed on this machine, and the GNOME check is not run here. It is done on a GNOME machine instead.

