---
title: SBOM
weight: 20
lead: A list of every component of your product, in CycloneDX 1.6.
---

`shipward sbom` reads your source tree with [Syft](https://github.com/anchore/syft) and writes a CycloneDX 1.6 JSON
SBOM. It reads lock files: `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `packages.lock.json` (NuGet),
`composer.lock`, `go.mod`, Python, Maven, Rust and Ruby lock files, and more.

```sh
shipward sbom --version 1.4.0
shipward sbom --version 1.4.0 --artifact build/game.exe --artifact-url https://example.com/game.exe
shipward sbom -o -                 # print it
```

With `--artifact`, the file you deliver gets its name and SHA-512 in the SBOM. `--form` says what that file is, for
the BSI fields: a list of `executable`, `bundle` (the file holds other files) and `structured`, or `none`. It overrides the product's settings in
`shipward.yaml`, so each download of the same release can have its own SBOM:

```sh
shipward sbom --version 1.4.0 --artifact dist/game-setup.exe --form executable,bundle -o sbom/setup.cdx.json
shipward sbom --version 1.4.0 --artifact dist/game.zip --form bundle -o sbom/zip.cdx.json
```

## What shipward fills in

Only what local files state as facts. It looks nothing up online.

| Field | Where it comes from |
|---|---|
| Dependencies | Every component gets a dependencies entry, and the SBOM states how complete the graph is: "unknown", because a scan cannot prove it found everything |
| Declared licences | Package metadata |
| Concluded licences | Only `licences.yaml`, which a person checks ([Licences](/docs/licences/)) |
| SHA-512, download URL, file name | npm, yarn and pnpm lock files; for NuGet, the restored `.nupkg` files |
| Creator, source URL | NuGet's `.nuspec` files |
| File properties | Registry packages are package files; your own .NET projects are executable files |

## .NET

Syft reads `packages.lock.json`, not `.csproj`. Turn lock files on once with
`<RestorePackagesWithLockFile>true</RestorePackagesWithLockFile>` in `Directory.Build.props`, run `dotnet restore`
and commit the lock files. Run `dotnet restore` before `shipward sbom`: hashes and creators come from the restored
packages (`$NUGET_PACKAGES` or `~/.nuget/packages`). The lock file's own `contentHash` is not used, because for
signed packages it is not the hash of the file.

## Components no package manager knows

A game engine, a copied add-on or vendored code goes into `components.extra.yaml`:

```yaml
components:
  - name: godot
    version: 4.7.2-stable
    type: framework
    purl: pkg:github/godotengine/godot@4.7.2-stable
    supplier: Godot Foundation
    creator: https://godotengine.org
    source: https://github.com/godotengine/godot/tree/4.7.2-stable
    licenses: [MIT]
    executable: true
```
