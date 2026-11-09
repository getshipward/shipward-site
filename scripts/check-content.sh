#!/usr/bin/env bash
# Rules for what may appear on the site, checked in the sources and in the built site (public/; run scripts/build.sh
# first). Usage: scripts/check-content.sh phone|wording
#
#   phone    No phone number anywhere: contact is by e-mail only. Fails on "tel:" and "+386".
#   wording  Shipward's wording rules: it "helps you prepare". Never "compliant", "certified", "malware", and no
#            "archive" wording (also in Slovenian: "arhiv", "hramba"): Shipward is a security tool.
set -euo pipefail

cd "$(dirname "$0")/.."

sources=(content layouts static assets hugo.toml)
[[ -d public ]] || { echo "check-content.sh: no public/; run scripts/build.sh first" >&2; exit 1; }

fail() {
  echo "$1" >&2
  exit 1
}

case "${1:-}" in
  phone)
    if grep -rnIi -e 'tel:' -e '+386' "${sources[@]}" public; then
      fail "check-content.sh: a phone number (or a tel: link) is above. Contact is by e-mail only."
    fi
    echo "phone: no phone numbers"
    ;;
  wording)
    if grep -rnIiE '\b(compliant|compliance-ready|certified|malware|archiv|arhiv|hramb)' content layouts public \
      --include='*.md' --include='*.html' --include='*.txt' --include='*.xml'; then
      fail "check-content.sh: a word Shipward must not use is above (see the top of this script)."
    fi
    echo "wording: ok"
    ;;
  *)
    fail "usage: scripts/check-content.sh phone|wording"
    ;;
esac
