#!/bin/sh

set -eu

apt_root=${APTUI_APT_ROOT:-/etc/apt}

for source_file in \
  "$apt_root/sources.list" \
  "$apt_root"/sources.list.d/*.list \
  "$apt_root"/sources.list.d/*.sources
do
  [ -f "$source_file" ] || continue
  sed -i 's|https://mexirica.github.io/aptui|https://rodrigomecheri.dev/aptui|g' "$source_file"
done