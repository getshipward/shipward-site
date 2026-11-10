---
title: Vulnerability scan
weight: 50
lead: Every component version against public advisories, matched on your machine.
---

```sh
shipward scan                          # the SBOM from shipward.yaml
shipward scan --format json file.json  # or markdown
shipward scan --offline                # no requests: the cached copy only
```

For each advisory that names a component version, `shipward scan` shows the advisory and its other IDs, the rating, the
fixed versions, whether it is known to be exploited, its EPSS score, and your decision. Advisories you marked
`affected` come first, then the ones you have not decided (open or under investigation), then the closed ones. Within
each group, exploited ones are at the top, then the highest EPSS score.

## Where the data comes from

`shipward` downloads a [signed snapshot](/docs/data/) of public advisories, made once a day, and only the files for the
ecosystems in your SBOM. It checks the signature and the SHA-256 of each file, keeps a copy in
`~/.cache/shipward`, and never accepts a copy older than the one it already has. The matching happens on your
machine: no package name is sent anywhere.

Two feeds come from their publishers directly, because their terms do not clearly allow us to pass them on: ENISA's
list of exploited vulnerabilities (EUVD) and FIRST's EPSS scores. `shipward` downloads both itself and keeps them for 20
hours.

## What a finding means

An advisory that names a version does not prove your product is affected. The vulnerable code may never run in your
product. A person decides that, and records it as a [decision](/docs/decisions/).

## In CI

```sh
shipward scan --summary "$GITHUB_STEP_SUMMARY"   # a table in the job summary
shipward scan --fail-on exploited                # exit 3 when an undecided advisory is known to be exploited
shipward scan --fail-on open                     # exit 3 when any advisory has no final decision
```

`--fail-on` is off by default. Shipward reports; you decide what stops a release.
