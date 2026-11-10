---
title: GitHub Action
weight: 70
lead: The SBOM, its field check, the vulnerability scan and the VEX export on every push, with the results in the job summary.
---

```yaml
jobs:
  shipward:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<commit>
      - run: dotnet restore --locked-mode       # .NET only: hashes and creators come from restored packages
      - uses: getshipward/shipward@<commit>     # pin a commit
        with:
          fail-on: none                         # or open, exploited, unchecked
```

The Action builds `shipward` from Shipward's source at the commit you pin, then runs `shipward sbom`, `shipward check sbom`,
`shipward scan` and `shipward triage export`. It does not run `shipward init`, `shipward licences` or `shipward report`
(the licences you confirmed in `licences.yaml` still go into the SBOM). The results:

- the job summary: the SBOM fields table and the vulnerability table with your decisions;
- the outputs `open` (advisories without a final decision), `exploited` (of those, the ones known to be
  exploited) and `unchecked` (components that could not be checked: see [What it can check](/docs/scan/#what-it-can-check));
- a workflow artifact with the `sbom/` folder: the SBOM and the VEX document.

## Keep each release's SBOM and VEX

Workflow artifacts expire (on GitHub after 90 days, unless your settings say otherwise). The SBOM and the VEX
document of each release belong to its technical documentation, which you keep for at least ten years after the
product is placed on the market, or for the support period if that is longer (CRA Article 13(13); the SBOM is Annex
VII point 8). Shipward does not keep them for you. Put them with the release, for example as release assets:

```yaml
    permissions:
      contents: write                     # to upload release assets
    steps:
      # ... checkout and the Shipward step, as above ...
      - if: startsWith(github.ref, 'refs/tags/')
        env: { GH_TOKEN: "${{ github.token }}" }
        run: gh release upload "$GITHUB_REF_NAME" sbom/sbom.cdx.json sbom/vex.openvex.json --clobber
```

The release must exist when this step runs (create it in an earlier step, or run the workflow on `release`).

## More than one delivered file

`--artifact` takes one file, so make one SBOM per download. The output `shipward` is the path of the binary the Action
built; later steps can run it:

```yaml
      - id: shipward
        uses: getshipward/shipward@<commit>
        with: { check: "false", scan: "false", export-vex: "false", upload: "false" }
      - env: { SHIPWARD: "${{ steps.shipward.outputs.shipward }}" }
        run: |
          "$SHIPWARD" sbom --version 1.4.0 --artifact dist/game-setup.exe --form executable,bundle -o sbom/setup.cdx.json
          "$SHIPWARD" sbom --version 1.4.0 --artifact dist/game.zip --form bundle -o sbom/zip.cdx.json
```

## Report deadlines

The Action does not look at your [report cases](../reports/). To hear about a missed Article 14 deadline, run
`shipward report --fail-on overdue` on a schedule. Run it only in a private repository that holds your cases: they
describe vulnerabilities that may not be fixed or public yet.

```yaml
name: CRA report deadlines
on:
  schedule:
    - cron: "17 * * * *"                      # every hour: the first deadline is 24 hours
  workflow_dispatch:
jobs:
  deadlines:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<commit>
      - id: shipward
        uses: getshipward/shipward@<commit>   # pin a commit
        with: { check: "false", scan: "false", export-vex: "false", upload: "false" }
      - env: { SHIPWARD: "${{ steps.shipward.outputs.shipward }}" }
        run: '"$SHIPWARD" report --fail-on overdue'
```

The job fails (exit code 3) while a report is overdue. GitHub sends the failure to the person who last changed the
schedule, if that person has Actions notifications turned on. GitHub can delay scheduled runs, and at times of high
load it can drop some ([GitHub's docs on `schedule`](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#schedule)).
So this job is a backstop, not your only reminder: keep the deadlines in your own calendar too.

## Inputs

| Input | Default | What |
|---|---|---|
| `working-directory` | `.` | Folder with `shipward.yaml` |
| `config` | `shipward.yaml` | Settings file |
| `version` | tag name, else the commit | Product version in the SBOM |
| `check` | `true` | List missing SBOM fields (never fails the job) |
| `scan` | `true` | Scan for known vulnerabilities |
| `fail-on` | `none` | `open`, `exploited` or `unchecked` to fail the job; several with commas (`exploited` also fails when CISA's or ENISA's list could not be loaded) |
| `export-vex` | `true` | Write the OpenVEX document when there are decisions |
| `artifact`, `artifact-url` | | The delivered file and its download URL |
| `upload`, `upload-name`, `upload-path` | `true`, `shipward`, `sbom` | The workflow artifact |

The Action sets up Go to build `shipward`, which changes the job's Go version.
