---
title: Security and privacy
weight: 90
lead: What shipward sends, what it keeps, and how to tell us about a problem.
---

## What shipward sends

Only plain downloads, the same for everyone:

| From | What | When |
|---|---|---|
| `data.shipward.dev` | The snapshot index and the files for the ecosystems in your SBOM | `shipward scan`, when the cached copy is older than the published one |
| `euvdservices.enisa.europa.eu` | ENISA's list of exploited vulnerabilities | At most every 20 hours |
| `epss.empiricalsecurity.com` | FIRST's EPSS scores | At most every 20 hours |

No package names, no SBOM, no findings and no decisions leave your machine. `shipward scan --offline` sends nothing at
all. `shipward sbom`, `shipward check sbom`, `shipward licences` and `shipward triage set` work offline.

## What shipward keeps

The snapshot and the two feeds, in `~/.cache/shipward` (your system's cache folder). Everything else it writes is in
your repository, where you put it.

## Reporting a vulnerability in Shipward

E-mail [daniel@sori-resitve.si](mailto:daniel@sori-resitve.si). See also
[security.txt](/.well-known/security.txt).
