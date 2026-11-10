---
title: Reports
weight: 65
lead: The reports the CRA asks for when a vulnerability is exploited or a severe incident happens, with their deadlines.
---

Since 11 September 2026, a manufacturer reports an actively exploited vulnerability in its product, and a severe
incident that affects the product's security, through ENISA's single reporting platform to its CSIRT coordinator
(CRA Article 14). The clock starts when you become aware of it.

| Report | Due |
|---|---|
| Early warning | 24 hours after becoming aware |
| Notification | 72 hours after becoming aware |
| Final report, vulnerability | 14 days after a fix or mitigation is available |
| Final report, severe incident | One month after the notification was submitted |
| Notice to affected users | In a timely manner |

`shipward report` keeps track of each case and writes drafts with the contents the article lists. You submit them on the
platform yourself: it has no API.

```sh
shipward report new vulnerability --title "Exploited flaw in login" --aware "2026-11-06 14:00" --advisory CVE-2026-1234
shipward report                         # every open case: the next report, its deadline, the time left
shipward report done CASE early-warning # after you submitted it
shipward report fixed CASE              # a fix is available: the final report is due 14 days later
shipward report drafts CASE --force     # write the drafts again after you changed report.yaml
```

## Times and deadlines

- A time is RFC 3339 (`2026-11-06T14:00:00+01:00`) or `2026-11-06 14:00` in your local time, on the command line
  and in `report.yaml` alike. A date without a time is refused.
- `shipward report new` writes your time zone into `report.yaml` (`zone: Europe/Ljubljana`). Times without an
  offset are read in that zone, and months are counted in it, so your computer and a CI runner give the same
  deadlines. Without `zone`, a time without an offset is read in the computer's zone, and a month keeps the offset of
  the time it starts from.
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

## A case

Each case is a folder in `cra/reports/`: `report.yaml` with the facts and the times, and four drafts:
`early-warning.md`, `notification.md`, `final.md` and `users.md`. Add what you learn to `report.yaml` (measures taken,
measures users can take, affected versions) and write the drafts again.

> A case describes something that may not be public yet. Keep it out of public repositories. shipward writes the files so
> that only you can read them.

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
does). Run it on a schedule, so a missed deadline
fails a job and you hear about it.
