---
title: SBOM fields
weight: 30
lead: Which fields Germany's BSI asks for, and which ones your SBOM misses.
---

The CRA asks for an SBOM in a common machine-readable format that covers at least the top-level dependencies. It does
not name the fields. Germany's Federal Office for Information Security (BSI) does, in its Technical Guideline
TR-03183-2, version 2.1.0 of 20 August 2025. `shipward check sbom` checks an SBOM against it:

```sh
shipward check sbom                    # the SBOM from shipward.yaml
shipward check sbom --all file.json    # name every component that misses a field
```

For each rule it shows how many components meet it, and names the ones that do not. It exits with code 1 when a
required field is missing, so you can use it in CI. Components inside other components are checked too.

It reads CycloneDX JSON, the format `shipward sbom` writes. For an SPDX or XML file it says what the file looks like
instead.

## The rules

| Level | Fields |
|---|---|
| Required for the SBOM | CycloneDX 1.6 or later, the creator's e-mail or URL, a timestamp, and no vulnerability information (that goes in a separate document, such as [VEX](/docs/decisions/)) |
| Required for each component | Creator, name, version, file name, dependencies and how complete they are, the distribution licence, SHA-512 of the deployable file, and three properties: whether it is executable, whether it contains other files, and whether its contents can still be listed |
| Required when the information exists | The SBOM's URI, each component's source code URI, download URI, other identifiers (purl, CPE), and its original licence |

A field counts only when its value is usable: a SHA-512 must be 128 hexadecimal digits, and a licence must be an SPDX
expression. That is an ID from the SPDX licence list (`MIT`), an expression (`MIT OR Apache-2.0`), or `LicenseRef-...`
for any other licence. Text such as "All rights reserved" does not count. `shipward sbom` refuses a `product.licence`
in `shipward.yaml`, or a `concluded` value in `licences.yaml`, that is not SPDX.

> The check lists missing fields. It is not a certification and not legal advice.

BSI may publish a newer version of TR-03183-2. The output names the version and date it checks against, and reminds
you to look for a newer one on [bsi.bund.de](https://www.bsi.bund.de).

Some fields can only exist after a build, such as the hash of the file you deliver. Pass that file with
`shipward sbom --artifact`.
