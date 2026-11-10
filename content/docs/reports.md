---
title: Reports
weight: 65
lead: The reports the CRA asks for when a vulnerability is exploited or a severe incident happens, with their deadlines.
---

Since 11 September 2026, a manufacturer reports an actively exploited vulnerability in its product, and a severe
incident that affects the product's security, through ENISA's single reporting platform to its CSIRT coordinator
(CRA Article 14). The clock starts when you become aware of it.

This also applies to products placed on the market before 11 December 2027 (Article 69(3)). The other CRA duties
reach those products only after a substantial modification made on or after that date (Article 69(2)). Reporting
also goes on after a product's support period has ended (Commission guidance C(2026) 5252, paragraph 210).

## Before you need it

Read this before the first case: in the first hours there is no time.

- **The platform.** Reports go through ENISA's single reporting platform,
  [portal.cra-srp.enisa.europa.eu](https://portal.cra-srp.enisa.europa.eu/). It is a web form, with no API, in
  English only. Log in as "Assigned Representative".
- **An account.** Each person who submits needs a personal EU Login account with multi-factor authentication. Make
  it now ([EU Login](https://ecas.ec.europa.eu/cas/login)). A manufacturer has one primary and up to 20 secondary
  representatives. ENISA asks you to register on the platform only when you have something to report; that takes a
  few minutes when the EU Login account exists.
- **Your CSIRT.** You report to the CSIRT coordinator of the Member State of your main establishment: where decisions
  on your products' cybersecurity are mainly taken, or else where you have the most employees in the EU
  (Article 14(7)). Without an establishment in the EU, take the first that applies: the Member State of your
  authorised representative, of your importer, of your distributor, or with the most users of your products. With
  the wrong CSIRT, the report may be invalidated, and you must send it again to the right one. Write the CSIRT into
  `reports.csirt`.
- **If the platform is down.** Wait until it is back and submit then. You may also contact your CSIRT directly, but
  you must still submit on the platform.
- **Who does it.** Decide who opens a case, who submits, and who stands in for them.

ENISA's pages ([platform](https://www.enisa.europa.eu/topics/product-security/vulnerability-services/eu-incident-response-and-cyber-crisis-management/single-reporting-platform-srp),
[FAQ](https://www.enisa.europa.eu/topics/product-security/vulnerability-services/eu-incident-response-and-cyber-crisis-management/single-reporting-platform-srp/frequently-asked-questions))
have the current state; this list is from November 2026.

## When the clock starts

You report two kinds of event:

- **An actively exploited vulnerability** in your product: there is reliable evidence that a malicious actor has
  exploited it in a system without the owner's permission (Article 3(42)). A flaw in a component counts only when it
  is exploited in your product. A flaw that cannot be reached, or has not been exploited, in your product is not
  reportable, but you still handle it and tell the component's maker (Article 13(6); guidance paragraph 218). A flag
  in CISA's or ENISA's list of exploited vulnerabilities is a reason to look, not a case by itself.
- **A severe incident** that affects your product's security: it affects, or can affect, the product's ability to
  protect the availability, authenticity, integrity or confidentiality of sensitive or important data or functions;
  or it has led, or can lead, to malicious code in the product or in a user's systems (Article 14(5)).

The clock starts when you become aware. When you notice something suspicious, or someone tells you, assess it at
once. You are aware when that first assessment gives you a reasonable degree of certainty that the flaw is exploited
or that the incident has compromised your product's security (guidance paragraphs 213 and 214). Examples:

- A customer sends logs that show an attacker used a flaw in your login. You check, and it is real: aware now.
- A researcher reports a flaw, with no sign of use by an attacker: not a case yet. Handle it as a vulnerability.
- Your update server served a changed installer: a severe incident, from when you confirm it.

Exploitation you knew of before 11 September 2026 need not be reported (guidance paragraph 217).

## Deadlines

| Report | At the latest |
|---|---|
| Early warning | 24 hours after becoming aware; send it without undue delay |
| Notification | 72 hours after becoming aware; send it without undue delay |
| Intermediate report | Only when the CSIRT asks for one, by the time it sets (Article 14(6)) |
| Final report, vulnerability | 14 days after a fix or mitigation is available |
| Final report, severe incident | One month after the notification was submitted |
| Notice to affected users | In a timely manner, in parallel with the reports, from when you become aware (Article 14(8)) |

The 24 and 72 hours are the outer limits: the law asks for both reports "without undue delay". A notification or a
final report is not needed when its information has already been provided (Article 14(2), (4)).

## shipward report

`shipward report` keeps track of each case and writes drafts with the contents the article lists. You submit them on the
platform yourself: it has no API.

```sh
shipward report new vulnerability --title "Exploited flaw in login" --aware "2026-11-06 14:00" --advisory CVE-2026-1234
shipward report                         # every open case: the next report, its deadline, the time left
shipward report done CASE early-warning # after you submitted it
shipward report done CASE users         # after you told the users
shipward report fixed CASE              # a fix is available: the final report is due 14 days later
shipward report intermediate CASE --due "2026-11-20 12:00"   # the CSIRT asked for an intermediate report
shipward report done CASE notification --skipped "already provided in the early warning"
shipward report drafts CASE --force     # write the drafts again after you changed report.yaml
```

## Times and deadlines

- A time is RFC 3339 (`2026-11-06T14:00:00+01:00`) or `2026-11-06 14:00` in your local time, on the command line
  and in `report.yaml` alike. A date without a time is refused.
- When Shipward can find your time zone's name, `shipward report new` writes it into `report.yaml`
  (`zone: Europe/Ljubljana`); on Windows it cannot, so add the line yourself. Times without an offset are read in that
  zone, and months are counted in it, so your computer and a CI runner give the same deadlines. Without `zone`, a
  time without an offset is read in the computer's zone, and a month keeps the offset of the time it starts from.
- One month ends on the same day of the next month. When that day does not exist, it ends on the last day of the
  month: a notification sent on 31 January gives 28 February (29 in a leap year). Shipward keeps the clock time of
  the notification, so its date is never later than the legal one; the legal end of the period can be later (the end
  of that day, or the next working day).
- Until you record the notification of an incident, the final report shows "at the latest": one month after the
  notification's 72-hour deadline. The date moves earlier when you record the notification.
- `shipward report fixed CASE --at TIME` can be before you became aware: the flaw may be fixed in a newer release
  already. The final report is still due 14 days after the fix. When that is before the notification is due,
  Shipward asks you to send the notification and the final report together by then. When the 14 days ended before
  you became aware, it asks for the final report with the notification.
- `shipward report done` records the reports in order: early warning, notification, final report. If you sent two
  together, record both with the same time. Times in the future, and times before you became aware, are refused.
- `--skipped REASON` records a notification or final report you did not send because its information was already
  provided. Shipward counts it as done and shows the reason. For an incident with a skipped notification, the final
  report's month counts from the early warning, not from when you recorded the skip.
- The notice to users is not in that order. `shipward report` lists it first, and reminds you of it after each
  report until you record it.

## The notice to users

Tell the affected users, and where appropriate all users, about the vulnerability or incident and what they can do
(Article 14(8)). Tell them as soon as you can: do not wait for the final report.

- Tell the affected users first, where they will see it: in the product, or by e-mail.
- Before a fix exists, say what users should do, not how the flaw works. Details can help attackers.
- Publish the full details once the security update is out (Annex I Part II(4)).

`users.md` is a draft of the notice. Where it fits, also give it in a structured, machine-readable form
(Article 14(8)): `shipward report drafts CASE --csaf` writes it as a
[CSAF 2.0](https://docs.oasis-open.org/csaf/csaf/v2.0/os/csaf-v2.0-os.html) document, `users.csaf.json`, from
`report.yaml`. It needs `versions`, `user_measures` and `manufacturer.url`. It is a draft: check it before it goes
out.

## The drafts follow the form

Each draft lists the fields its stage of the form has, in the form's order, with ENISA's field numbers (from the
[CRA SRP Glossary](https://www.enisa.europa.eu/topics/product-security/single-reporting-platform-srp/cra-srp-glossary2),
version 1.6) and the length limits. The awareness time is shown in UTC, as the form asks. A value that is too long for
its field is marked. Check the fields against the form: ENISA can change it.

Add what you know to `report.yaml`, then `shipward report drafts CASE --force`:

```yaml
summary: An attacker used a flaw in the login of Example Game 1.3 to take over accounts. A fix is being prepared.
occurred: 2026-11-04 22:00          # when it was exploited or happened, as far as known
versions: [1.3.0, 1.3.1]
product_type: default               # default, important or critical (Annexes III and IV)
mitigation_expected: "yes"
attack_vector: Network, through the login form; no account needed.
severity: High. ...
impact: Account takeover of any player. ...
sensitivity: Not public yet; details help attackers until the update is out.
```

## A case

Each case is a folder in `cra/reports/`: `report.yaml` with the facts and the times, and four drafts:
`early-warning.md`, `notification.md`, `final.md` and `users.md`, plus `intermediate.md` when the CSIRT asks for an
intermediate report. Add what you learn to `report.yaml` (measures taken,
measures users can take, affected versions) and write the drafts again.

> A case describes something that may not be public yet. Keep it out of public repositories. shipward writes the files so
> that only you can read them. When `shipward report new` creates the `cra/reports/` folder, it writes
> `cra/reports/.gitignore` with `*`, so git ignores the cases. When git does not ignore a case and the repository has a
> remote, `shipward report new` warns you.

## Settings

```yaml
product:
  markets: [EU]          # Member States where the product is available; EU means all of them
reports:
  csirt: SI-CERT         # the CSIRT coordinator of the Member State of your main establishment
  dir: cra/reports
```

## In CI

`shipward report --fail-on overdue` exits with code 3 when a report is overdue (`--fail-on none`, the default, never
does). Run it on a schedule, so a missed deadline fails a job and you hear about it. The
[GitHub Action](../github-action/) page has an example.

The job needs the cases, so run it only in a private repository that keeps them. Two ways:

- Keep the cases in a private repository of their own, with a small `shipward.yaml` (product, manufacturer and
  `reports`). Delete `cra/reports/.gitignore` there, commit the cases, and run the job in that repository.
- If the product's repository is private, delete `cra/reports/.gitignore` there and commit the cases. Do not do this
  in a public repository.
