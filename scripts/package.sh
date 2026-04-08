#!/usr/bin/env bash

set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
example_root="$root/examples/service-template"
source_root="$example_root/source"
dist_root="$example_root/dist"
staging_root="$dist_root/echo-service-win32"
archive_path="$dist_root/echo-service-win32.zip"

if [[ ! -d "$source_root" ]]; then
  echo "missing example source directory: $source_root" >&2
  exit 1
fi

rm -rf "$staging_root" "$archive_path"
mkdir -p "$dist_root" "$staging_root"
cp -R "$source_root"/. "$staging_root"/

python3 - <<'PY' "$staging_root" "$archive_path"
import pathlib
import sys
import zipfile

source = pathlib.Path(sys.argv[1])
archive = pathlib.Path(sys.argv[2])
archive.parent.mkdir(parents=True, exist_ok=True)

with zipfile.ZipFile(archive, "w", compression=zipfile.ZIP_DEFLATED) as zf:
    for path in sorted(source.rglob("*")):
        if path.is_file():
            zf.write(path, path.relative_to(source))
PY

echo "Created example artifact: $archive_path"
