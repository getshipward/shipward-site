---
title: Advisories
weight: 62
lead: A draft advisory about each fixed vulnerability, for people and in CSAF for tools.
---

Once a security update is available, the CRA asks you to publish information about the vulnerabilities it fixes
(Annex I Part II(4)). Publish it with the update (Part II(8)). This is for every fixed vulnerability in your product,
not only for the exploited ones you [report](../reports/). The advisory names:

- what the vulnerability is;
- the product and the versions affected;
- the impact;
- the severity;
- clear steps users can follow to fix it, for example the version to update to.

In duly justified cases you may wait to publish until users could install the update (Part II(4)).

```sh
shipward advisory new CVE-2026-1234 --affected 1.3.0 --affected 1.3.1 --fixed-in 1.4.0 --severity high
shipward advisory new EXAMPLE-2026-001 --title "Saved games could be read by other users" --csaf
```

`shipward advisory new ID` writes `cra/advisories/ID.md` (in your `docs.dir`). It fills in what it knows: your
product and contact from `shipward.yaml`, the component, the fixed version and the action from your decision in
`vex.yaml`, and the summary, other IDs, rating and link from the cached vulnerability snapshot, when the advisory names
a component of your SBOM. Everything else is a TODO. It does not overwrite a draft that exists; `--force` does.

The ID can be a CVE, GHSA or EUVD ID, or your own for a flaw in your own code.

| Flag | What |
|---|---|
| `--title` | A short title |
| `--affected` | An affected product version; repeat it for more |
| `--fixed-in` | The product version with the security update |
| `--component` | The affected component (a package URL), when the flaw is in a dependency |
| `--severity` | For example `high` |
| `--description`, `--impact`, `--remediation` | The texts; the remediation defaults to "Update to version X or later" |
| `--csaf` | Also write `ID.csaf.json` |
| `-o` | Another output folder |

> Until the update is out, the draft describes an unfixed flaw. Keep it out of public repositories until you publish
> it.

## CSAF

With `--csaf`, Shipward also writes a [CSAF 2.0](https://docs.oasis-open.org/csaf/csaf/v2.0/os/csaf-v2.0-os.html)
document with the profile "Security advisory". Users and their tools can read it automatically. It needs the affected
versions, the fixed version and `manufacturer.url`. Its tracking status is `draft` and its version is `0`. When you
publish it, set the status to `final`, the version to `1` (with a revision history entry numbered `1`) and the dates.

## Record it

A `fixed` decision in `vex.yaml` is final for VEX, but the CRA work is done only when the update and the advisory
are out. Record both in the decision:

```sh
shipward triage set CVE-2026-1234 pkg:nuget/Lib@2.0.0 --status fixed --fixed-in 1.4.0 \
  --advisory https://example.com/security/CVE-2026-1234
```

Until both are recorded, `shipward triage` lists the decision under "Fixed, but the update or the advisory is not
recorded yet". It does not change `--fail-on`.
