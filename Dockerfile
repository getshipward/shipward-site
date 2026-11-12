# The shipward.dev image: the built Hugo site and the small Go server that serves it (server/), on Chainguard's static base
# (no shell, no package manager), running as the nonroot user (65532). Plain HTTP on 8080; TLS ends at the gateway.
# It writes nothing to disk (the site is read into memory at startup), so a read-only root file system is fine.
#
#   docker build -t shipward-site . && docker run --rm -p 8080:8080 shipward-site
#
# Images are pinned by digest; Renovate (.github/renovate.json5) proposes updates, and moves the golang image together
# with the go line in server/go.mod.

# 1. The site, with scripts/build.sh: it downloads the Hugo release named in .hugo-version and checks its checksum.
FROM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS site
WORKDIR /src
COPY . .
RUN scripts/build.sh && test -f public/index.html

# 2. The server.
FROM golang:1.27.4-trixie@sha256:b54b2639de8c1bfe538ec5ffc45f29b1932235ac5f6d63c059853646677112d9 AS server
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY server/go.mod ./
COPY server/*.go ./
RUN go build -trimpath -ldflags="-s -w" -o /site-server .

FROM cgr.dev/chainguard/static:latest@sha256:9e2306b457ff71ce370043dc68c7b52bb1e6622c1c6c6714edbed57496f3116e
COPY --from=server /site-server /site-server
COPY --from=site /src/public /site
USER 65532:65532
ENV LISTEN_ADDR=:8080 SITE_DIR=/site
EXPOSE 8080
ENTRYPOINT ["/site-server"]
