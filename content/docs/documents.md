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
| `cra/simplified-declaration-of-conformity.md` | The short form you may give with the product instead: two fixed sentences and the address of the full declaration | CRA Annex VI |
| `cra/technical-documentation.md` | The outline of the technical documentation, with the risk assessment: its context, a table for each requirement of Annex I, Part I (with a column for why one does not apply) and one for each vulnerability handling requirement of Part II, with the files Shipward makes, where it makes one | CRA Annex VII, Annex I, Article 13(3) and (4) |

Some drafts hold choices only you can make, so they stay TODO until you make them:

- `SECURITY.md` promises response times: when you confirm a report, when you say what happens next, and how long you
  ask a reporter to keep it private. Set them in `shipward.yaml` (below), or write them into the file.
- `SECURITY.md` says which versions get security updates. Fill in the table, or say that only the latest version
  gets them. Article 13(10) allows that for later, substantially modified versions, only if the users of the earlier
  versions can get the latest version free of charge and have no additional costs to adjust their hardware and
  software environment.
- `security.txt` must have a `Contact` field (RFC 9116). Without `security.contact` or `manufacturer.email`, the field
  is a TODO and `shipward init` warns. The draft's `Expires` is 11 months from the day you run `shipward init`: write
  the file again before then.
- The support period: `shipward init` refuses a `support_until` month that is over. It warns when the support period
  ends less than five years after `placed_on_market` (or after today, when that is not set): Article 13(8) asks for at
  least five years, unless the product is expected to be in use for less time. The drafts also hold TODOs for showing
  the end date at the time of purchase (Article 13(19)) and for keeping each security update available for at least
  ten years (Article 13(9)).

> These are drafts that help you prepare. They are not legal advice and do not show conformity. Every file starts
> with a short note saying so; remove it once you have filled in the TODOs and checked every sentence. Sign the
> declaration only when the technical documentation is complete.

## Product class and the declaration

The conformity assessment you may use depends on the product's class (Article 32). Set it as `product.class` once you
have checked the product against Annex III and Annex IV: the technical documentation draft has a short checklist, and
the technical descriptions of the categories are in Commission Implementing Regulation (EU)
[2025/2392](https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32025R2392).

| `product.class` | What point 7 of the declaration says |
|---|---|
| not set | TODO, and the notified body is TODO too |
| `default` (not in Annex III or IV) | Notified body: none. Internal control (module A) |
| `important-1` (Annex III, class I) | TODO. Internal control only if you applied harmonised standards, common specifications or a certification scheme in full; otherwise a notified body (modules B and C, or H) |
| `important-2` (Annex III, class II) | TODO. A notified body (modules B and C, or H) or a certification scheme; never internal control alone |
| `critical` (Annex IV) | TODO. A certification scheme where one is required, otherwise as for class II |

Free and open-source software in Annex III may use internal control if its technical documentation is public
(Article 32(5)); the draft says so for those classes.

The declaration's point 3 says "manufacturer", where Annex V says "provider": the manufacturer draws up the
declaration and takes responsibility for it (Articles 13(12) and 28). Make the declaration available in the languages
required by each Member State where the product is placed on the market or made available, also when you supply it
free of charge in the course of a commercial activity (Article 28(2), Article 3(22)).

For software, the CE marking goes on the declaration of conformity or on the product's website, and only once the
conformity assessment shows that the requirements are met (Articles 13(12) and 30). Set `product.declaration_url` to
the address of the full declaration: the simplified declaration and the user information link to it.

## Settings it uses

```yaml
product:
  name: Example Game
  description: an online board game for Windows, Linux and Android
  purpose: Players play online with friends or bots.
  website: https://example.com/game
  support_until: "2031-12"        # the support period runs to the end of this month
  placed_on_market: "2026-12-01"  # when it is (or was) placed on the market
  class: default                  # default, important-1, important-2 or critical (see above)
  declaration_url: https://example.com/game/declaration   # where the full EU declaration of conformity is
manufacturer:
  name: Example Ltd
  address: "Street 1, 1000 City, Country"
  email: security@example.com
security:
  contact: mailto:security@example.com   # default: the manufacturer's e-mail
  policy: https://example.com/security   # where SECURITY.md is published
  languages: [en]
  acknowledge_working_days: 5            # SECURITY.md: you confirm a report within 5 working days
  assessment_days: 30                    # you say within 30 days whether you can reproduce it
  disclosure_days: 90                    # you ask the reporter to keep it private for up to 90 days
docs:
  dir: cra                             # where the drafts go (default cra); SECURITY.md goes to the repository root
```

All of these are optional except the two names. What is missing becomes a TODO.
