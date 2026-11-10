---
title: Licences
weight: 40
lead: Declared licences come from package metadata; the ones you confirm come from you.
---

An SBOM can carry two kinds of licence for a component. The declared licence is what its creator assigned. The
concluded licence (the BSI's "distribution licence") is the one under which you can pass it on, and a person decides
it. `shipward` never sets a concluded licence by itself.

```sh
shipward licences                  # components without a confirmed licence
shipward licences --write          # draft an entry for each component with one declared SPDX licence
```

`--write` adds drafts, marked `confirmed: false`. `shipward sbom` ignores a draft until a person checks the licence
and removes that line (or sets it to `true`). So the concluded licence in the SBOM is always one a person confirmed.
`shipward licences` lists the drafts that still wait for a check, and a second `--write` does not add them again.

```yaml
- component: pkg:nuget/MessagePack
  declared: MIT
  concluded: MIT
  confirmed: false # check the licence, then remove this line
```

`concluded` must be an SPDX expression: an ID such as `MIT`, an expression such as `MIT OR Apache-2.0`, or
`LicenseRef-...` for a licence that is not on the SPDX list.

After the check, remove the `confirmed` line, commit `licences.yaml` and run `shipward sbom` again. A component whose
declared licence is missing, not SPDX, or one of several gets no draft: find its licence and add the entry yourself.

An entry without `@version` applies to every version while the declared licence stays the same. When a new version
declares another licence, the entry no longer applies, and `shipward licences` lists the component again.
