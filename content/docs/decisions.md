---
title: Decisions (VEX)
weight: 60
lead: Record whether an advisory affects your product, and tell your users.
---

VEX (Vulnerability Exploitability eXchange) says whether a known vulnerability affects a product. Shipward keeps your
decisions in `vex.yaml` in your repository, and writes them as an [OpenVEX](https://openvex.dev) document for your
users and their scanners.

```sh
shipward triage                    # advisories without a final decision
shipward triage set GHSA-5crp-9r3c-p9vr pkg:nuget/Newtonsoft.Json@12.0.1 --status not_affected \
  --justification vulnerable_code_not_in_execute_path --impact "We never parse untrusted JSON." --by alex
shipward triage export             # sbom/vex.openvex.json
```

`shipward triage` prints a ready `shipward triage set` line for each open advisory. A decision can name the advisory by any of
its IDs (GHSA or CVE). With `--all-versions` it applies to every version of the component.

## Statuses

The words are OpenVEX's.

| Status | Meaning | Needs |
|---|---|---|
| `not_affected` | The product is not affected | A justification or an impact statement |
| `affected` | The product is affected | An action statement: what users should do |
| `fixed` | A release fixed it | For the CRA: `fixed_in` and `advisory` once the [advisory](../advisories/) is published |
| `under_investigation` | Not decided yet | |

Justifications for `not_affected`: `component_not_present`, `vulnerable_code_not_present`,
`vulnerable_code_not_in_execute_path`, `vulnerable_code_cannot_be_controlled_by_adversary`,
`inline_mitigations_already_exist`.

## vex.yaml

```yaml
- id: GHSA-5crp-9r3c-p9vr
  component: pkg:nuget/Newtonsoft.Json@12.0.1
  status: not_affected
  justification: vulnerable_code_not_in_execute_path
  impact: We never parse untrusted JSON.
  by: alex
  date: "2026-11-06"
- id: CVE-2026-1234
  component: pkg:nuget/Lib@2.0.0
  status: fixed
  fixed_in: 1.4.0                          # the product version with the update
  advisory: https://example.com/security/CVE-2026-1234
  date: "2026-11-10"
```

Edit it by hand or with `shipward triage set`; comments are kept. `shipward triage` also lists decisions that no advisory needs
any more, for example after an update.

## The OpenVEX document

`shipward triage export` names your product with a link to its SBOM, and each component by its package URL. A decision for
every version is written for the versions your product ships. The same `vex.yaml` always gives the same document.
