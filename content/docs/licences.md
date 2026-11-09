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
shipward licences --write          # draft an entry for each component with one declared licence
```

Check each entry `--write` adds before you commit `licences.yaml`: by committing it, you confirm the licences. Then
run `shipward sbom` again.

```yaml
- component: pkg:nuget/MessagePack
  declared: MIT
  concluded: MIT
```

An entry without `@version` applies to every version while the declared licence stays the same. When a new version
declares another licence, the entry no longer applies, and `shipward licences` lists the component again.
