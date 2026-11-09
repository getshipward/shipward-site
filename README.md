# shipward.dev

The website of [Shipward](https://github.com/getshipward/shipward): what it is, and its documentation. A static Hugo
site with its own theme (no third-party theme), served by a small Go server in `server/` (headers and redirects from
`static/_headers` and `static/_redirects`, a 404 page, no request logs, no cookies).

## Writing rules

- Shipward "helps you prepare". Never "compliant", "certified" or "approved", never "malware", and no "archive"
  wording: Shipward is a security tool. CI checks the words (`scripts/check-content.sh wording`).
- No phone numbers: contact is by e-mail only (`scripts/check-content.sh phone`).
- When Shipward's commands change, change the docs pages with them.

## Layout

| Path | What |
|---|---|
| `content/` | The pages, Markdown. Docs order: `weight` in each page's front matter |
| `layouts/index.html`, `layouts/partials/diagram.html` | The home page and its diagram |
| `layouts/`, `assets/` | The theme: templates, one stylesheet (Indigo and Cyan tokens at the top, light and dark), `theme.js` (in `<head>`) and `site.js` (theme toggle) |
| `static/` | Favicon (the logo), `.well-known/security.txt`, `_headers` (CSP, caching), `_redirects` |
| `server/` | The site server (Go, standard library only) |

## Build

```bash
scripts/build.sh            # into public/; downloads the Hugo version in .hugo-version
scripts/link-check.sh       # every internal link and #fragment
scripts/check-content.sh phone; scripts/check-content.sh wording
docker build -t shipward-site . && docker run --rm -p 8080:8080 shipward-site
```

Every push to `main` builds the image `ghcr.io/getshipward/shipward-site:main-N` (`.github/workflows/ci.yml`).
