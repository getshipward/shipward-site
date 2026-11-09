#!/usr/bin/env bash
# Checks every internal link and #fragment of the built site (public/) with htmltest. Run scripts/build.sh first.
set -euo pipefail

cd "$(dirname "$0")/.."

# renovate: datasource=github-releases depName=wjdp/htmltest
HTMLTEST_VERSION=0.17.0

dir=.cache/htmltest-$HTMLTEST_VERSION
if [[ ! -x $dir/htmltest ]]; then
  case "$(uname -m)" in
    x86_64 | amd64) arch=amd64 ;;
    aarch64 | arm64) arch=arm64 ;;
    *) echo "link-check.sh: no htmltest download for $(uname -m)" >&2; exit 1 ;;
  esac
  tarball=htmltest_${HTMLTEST_VERSION}_linux_${arch}.tar.gz
  base=https://github.com/wjdp/htmltest/releases/download/v$HTMLTEST_VERSION
  mkdir -p "$dir"
  curl -fsSL --retry 3 -o "$dir/$tarball" "$base/$tarball"
  curl -fsSL --retry 3 -o "$dir/checksums.txt" "$base/htmltest_${HTMLTEST_VERSION}_checksums.txt"
  (cd "$dir" && grep " $tarball\$" checksums.txt | sha256sum -c --quiet -)
  tar -xzf "$dir/$tarball" -C "$dir" htmltest
  rm -f "$dir/$tarball"
fi

mkdir -p .cache/htmltest
"$dir/htmltest" --conf .htmltest.yml
