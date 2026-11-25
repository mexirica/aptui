#!/bin/sh

set -eu

apt_root=${APTUI_APT_ROOT:-/etc/apt}

for source_file in \
  "$apt_root/sources.list" \
  "$apt_root"/sources.list.d/*.list \
  "$apt_root"/sources.list.d/*.sources
do
  [ -f "$source_file" ] || continue
  grep -Fq 'https://mexirica.github.io/aptui' "$source_file" || continue
  if ! sed --follow-symlinks -i 's|https://mexirica.github.io/aptui|https://rodrigomecheri.dev/aptui|g' "$source_file"; then
    printf 'aptui: warning: could not migrate repository URL in %s\n' "$source_file" >&2
  fi
done
