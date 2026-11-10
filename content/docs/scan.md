---
title: Vulnerability scan
weight: 50
lead: Every component version against public advisories, matched on your machine.
---

```sh
shipward scan                          # the SBOM from shipward.yaml
shipward scan --format json file.json  # or markdown
shipward scan --offline                # no requests: the cached copy only
shipward scan --max-age 3d             # exit 1 when the data is older than 3 days
```

For each advisory that names a component version, `shipward scan` shows the advisory and its other IDs, the rating, the
fixed versions, whether it is known to be exploited, its EPSS score, and your decision. Advisories you marked
`affected` come first, then the ones you have not decided (open or under investigation), then the closed ones. Within
each group, exploited ones are at the top, then the highest EPSS score.

## What it can check

`shipward scan` matches components by their package URL (purl), in these ecosystems: npm, NuGet, Packagist
(Composer), Go, PyPI, Maven, crates.io (Rust) and RubyGems. Any other component is **not checked**, and no advisory
is ever shown for it. That includes a game engine (`pkg:github/...`), C or C++ libraries, system packages, container
images and anything without a purl. Check those yourself, for example on the project's security page.

The first line of the output says how many components were not checked ("2 of 385 components not checked"), and
`--all` names them all. The same count is in the JSON (`not_checked`) and the job summary.
`--fail-on unchecked` exits 3 when any component could not be checked.

## Where the data comes from

`shipward` downloads a [signed snapshot](/docs/data/) of public advisories, made once a day, and only the files for the
ecosystems in your SBOM. It checks the signature and the SHA-256 of each file, keeps a copy in
`~/.cache/shipward`, and never accepts a copy older than the one it already has. The matching happens on your
machine: no package name is sent anywhere.

Two feeds come from their publishers directly, because their terms do not clearly allow us to pass them on: ENISA's
list of exploited vulnerabilities (EUVD) and FIRST's EPSS scores. `shipward` downloads both itself and keeps them for 20
hours.

An advisory counts as known to be exploited when CISA's catalog or ENISA's list names any of its IDs: the CVE, the
GHSA or other advisory ID, or ENISA's own EUVD ID.

## How old the data is

Every scan shows the age of the snapshot and of the ENISA and EPSS copies, in the text, the JSON
(`snapshot_age_hours`, `snapshot_expires` and `feeds`) and the job summary. It warns when:

- the snapshot or a feed copy is more than 48 hours old (for example with `--offline`, which uses copies of any age);
- the snapshot has expired: each one expires 7 days after it was made, so a server that keeps offering an old
  snapshot is noticed even on a CI runner that starts with an empty cache;
- a feed could not be loaded, or came back without one usable entry (CISA's and ENISA's lists are never empty, so
  such a list counts as not loaded, and an older good copy is kept). Then the scan says so, and lists only the sources it really used.

`--max-age` turns the age warning into a failure: `shipward scan --max-age 3d` exits 1 when the snapshot or a feed copy
is older than 3 days. It takes hours (`36h`) or days (`3d`).

## Exploited advisories and Article 14

CISA and ENISA list vulnerabilities that someone, somewhere, has exploited. That alone is not a case you must report.
Under the CRA you report an *actively exploited vulnerability* in your product: one with reliable evidence that a
malicious actor exploited it (Article 3(42), Article 14(1)). So for each exploited advisory without a final decision,
`shipward scan` and `shipward triage` ask you to decide whether it is actively exploited in your product, and print the
command that opens a report case if it is:

```text
Known to be exploited somewhere. That alone is not a case to report: decide whether it is actively exploited
in your product (CRA Art. 3(42), 14(1)). If it is, the 24-hour early warning may already be due:
  shipward report new vulnerability --title 'GHSA-... in Lib@1.0.0' --advisory GHSA-... --component pkg:nuget/Lib@1.0.0
```

`shipward triage set ... --status affected` prints the same when the cached snapshot lists the advisory as exploited.
When it cannot tell (no cached snapshot, no such advisory for that component, or CISA's or ENISA's list missing), it
says so and gives the same guidance.
The 24 hours run from when you became aware (see [Reports](/docs/reports/)).

These lists only cover public advisories. Exploitation of your own code never shows in them: if you learn of it from a
user, a researcher or your own logs, open a case yourself.

## What a finding means

An advisory that names a version does not prove your product is affected. The vulnerable code may never run in your
product. A person decides that, and records it as a [decision](/docs/decisions/).

## In CI

```sh
shipward scan --summary "$GITHUB_STEP_SUMMARY"   # a table in the job summary
shipward scan --fail-on exploited                # exit 3 when an undecided advisory is known to be exploited
shipward scan --fail-on open                     # exit 3 when any advisory has no final decision
shipward scan --fail-on unchecked                # exit 3 when a component could not be checked
shipward scan --fail-on exploited,unchecked      # several, with commas
```

`--fail-on` is off by default. Shipward reports; you decide what stops a release. With `--fail-on exploited`, a scan that
could not load CISA's or ENISA's list exits 1: it cannot tell whether an advisory is exploited, so it does not pass
quietly.
