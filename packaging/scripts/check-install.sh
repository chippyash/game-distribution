#!/bin/sh
# Checks an installed gamedist package. Run it after installing the .deb or .rpm, with desktop-file-utils present
set -e
gamedist help >/dev/null
test -x /usr/bin/gamedist
test -x /usr/bin/gamedistapp
test -f /usr/share/applications/gamedistapp.desktop
desktop-file-validate /usr/share/applications/gamedistapp.desktop
for size in 48x48 128x128 256x256; do
    test -f "/usr/share/icons/hicolor/$size/apps/gamedistapp.png"
done
if ldd /usr/bin/gamedistapp | grep -q "not found"; then
    ldd /usr/bin/gamedistapp | grep "not found"
    echo "gamedistapp is missing libraries"
    exit 1
fi
echo "install check passed"
