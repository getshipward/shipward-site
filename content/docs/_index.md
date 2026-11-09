---
title: Documentation
lead: How Shipward's command, shipward, works, from the first SBOM to the decisions you publish.
---

> Shipward is in use for our own products and is not available to others yet. These pages describe how it works.

Shipward is one command, `shipward`, and a GitHub Action that runs it. It works in your repository and on your machine or
CI runner. It needs no account and uploads nothing.

| Command | What it does |
|---|---|
| `shipward sbom` | Writes a CycloneDX 1.6 SBOM of your product |
| `shipward check sbom` | Lists the fields the SBOM misses for BSI TR-03183-2 |
| `shipward init` | Writes drafts of the CRA documents: security policy, security.txt, user information, declaration of conformity, technical documentation |
| `shipward licences` | Lists components without a confirmed licence |
| `shipward scan` | Matches the SBOM against public vulnerability data, locally |
| `shipward triage` | Records your decisions about advisories (VEX) and exports them as OpenVEX |
| `shipward report` | Drafts of the Article 14 reports and their deadlines |

Start with [Getting started](/docs/getting-started/).
