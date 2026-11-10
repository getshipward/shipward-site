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

Say who decided with `--by` (a name or role). It is optional, but without it nobody can trace the decision later, so
`shipward triage set` warns when it is missing. `--date` takes a day (`2026-11-06`, the default is today) or a full
RFC 3339 time (`2026-11-06T14:30:00+01:00`).

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

## Telling the component's maintainer

When you find a vulnerability in a component you use, the CRA asks you to report it to whoever maintains that
component, and to share your fix with them if you made one (Article 13(6)). Record what you did in `upstream`:

```yaml
- id: GHSA-5crp-9r3c-p9vr
  component: pkg:nuget/Newtonsoft.Json@12.0.1
  status: affected
  action: Update to 13.0.1.
  by: alex
  date: "2026-11-06"
  upstream:
    reported: "2026-11-07"
    link: https://example.com/newtonsoft-json/issues/4
    fix_shared: true
```

`shipward triage set ... --upstream-reported DATE --upstream-link URL --upstream-fix-shared` writes it. A later
`shipward triage set` on the same advisory and component keeps it, and each upstream flag changes only its own field.
If the advisory came from the maintainer, they know already: there is nothing to report. After an `affected` decision
without `upstream`, `shipward triage set` reminds you, and `shipward triage` shows "upstream: not recorded". `shipward scan` shows what you
recorded next to the decision. The OpenVEX document does not carry it.

## The OpenVEX document

`shipward triage export` names your product with an ID that stays the same for one release: the product's package URL
from the SBOM when it has one (with the product version added when it has none), else `pkg:generic/<name>@<version>`. Each component is named by its package URL. A
decision for every version is written for the versions your product ships, with its date or time in UTC.

Each release has one document, and its `@id` stays the same. When the document already exists, `shipward triage
export` compares the decisions. Unchanged decisions give an unchanged file. Changed ones raise `version` by one and
set `last_updated`. Keep the exported file (for example with the release) to keep the count: without it, the next
export starts at version 1 again, and says so on stderr.
