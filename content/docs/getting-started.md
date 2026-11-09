---
title: Getting started
weight: 10
lead: One settings file, then three commands.
---

## shipward.yaml

Put a `shipward.yaml` in the root of your product's repository:

```yaml
product:
  name: Example Game
  licence: LicenseRef-example-proprietary   # SPDX expression you distribute the product under
manufacturer:
  name: "Example Software Ltd"
  url: https://example.com
  email: security@example.com
sbom:
  exclude: ["./tests/**", "./**/bin/**", "./**/obj/**"]   # left out of the scan
```

Only `product.name` and `manufacturer.name` are required. The defaults: the SBOM goes to `sbom/sbom.cdx.json`, extra
components come from `components.extra.yaml`, confirmed licences from `licences.yaml`, and decisions from `vex.yaml`.

## The first run

```sh
shipward sbom --version 1.4.0      # sbom/sbom.cdx.json
shipward check sbom                # which SBOM fields are missing
shipward scan                      # known vulnerabilities in your components
```

Then record a decision for each advisory `shipward scan` shows ([Decisions](/docs/decisions/)), and add the
[GitHub Action](/docs/github-action/) so this runs on every push.

## Scan a clean checkout

Run `shipward sbom` on a clean checkout of your source, as CI does. Old build output in your working folder can contain
packages the product no longer ships, and they would appear in the SBOM.
