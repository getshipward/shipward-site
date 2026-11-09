---
title: CRA documents
weight: 15
lead: Drafts of the documents the CRA asks a manufacturer to have, filled in from shipward.yaml.
---

```sh
shipward init --product "Example Game" --manufacturer "Example Ltd" --email security@example.com   # no shipward.yaml yet: writes one too
shipward init                                   # all drafts
shipward init --only security.txt --force       # write one draft again
```

`shipward init` writes what `shipward.yaml` knows and marks everything else with a visible `TODO:`. It prints how many TODOs
each file has. It never overwrites a file unless you add `--force`.

| File | What it is | Follows |
|---|---|---|
| `SECURITY.md` | Your coordinated vulnerability disclosure policy: where to report, what happens next, supported versions | CRA Annex I, part II |
| `cra/security.txt` | The file security researchers look for, to publish at `/.well-known/security.txt` on your website | RFC 9116 |
| `cra/user-information.md` | The information users must get: manufacturer, contact for vulnerabilities, intended purpose, support period, instructions | CRA Annex II |
| `cra/declaration-of-conformity.md` | The EU declaration of conformity, unsigned | CRA Annex V |
| `cra/technical-documentation.md` | The outline of the technical documentation, with a table for each essential requirement | CRA Annex VII and Annex I, part I |

> These are drafts that help you prepare. They are not legal advice and do not show conformity. Every file starts
> with a short note saying so; remove it once you have filled in the TODOs and checked every sentence. Sign the
> declaration only when the technical documentation is complete.

## Settings it uses

```yaml
product:
  name: Example Game
  description: an online board game for Windows, Linux and Android
  purpose: Players play online with friends or bots.
  website: https://example.com/game
  support_until: "2031-12"        # end of the support period (at least five years, Article 13(8))
manufacturer:
  name: Example Ltd
  address: "Street 1, 1000 City, Country"
  email: security@example.com
security:
  contact: mailto:security@example.com   # default: the manufacturer's e-mail
  policy: https://example.com/security   # where SECURITY.md is published
  languages: [en]
docs:
  dir: shipward                        # where the drafts go; SECURITY.md goes to the repository root
```

All of these are optional except the two names. What is missing becomes a TODO.
