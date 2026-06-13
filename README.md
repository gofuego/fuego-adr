# fuego-adr

A documentation site generator for **Architecture Decision Records** (ADRs),
built on the [Fuego](https://github.com/FabioSol/fuego) meta-engine.

Write your decisions as `*.adr.md` files, point `fuego-adr` at the folder, and
get a styled, navigable site: a status dashboard, a chronological timeline, a
per-file "what decisions touched this?" index, tag pages, and a page per
decision — with section completeness and supersession links validated for you.

```bash
fuego-adr build docs/adr
```

fuego-adr ships as both a **CLI** (zero-config: point it at a folder) and an
importable **format pack** (`adr.Pack()`) you can drop into any Fuego project.

---

## Install

```bash
go install github.com/FabioSol/fuego-adr@latest
```

Requires Go 1.23+. The binary lands in `$GOPATH/bin` (usually `~/go/bin`); make
sure that's on your `PATH`. Or run without installing:

```bash
go run github.com/FabioSol/fuego-adr@latest build docs/adr
```

## Quick start

```bash
# Create your first decision record (auto-numbered, slug from the title)
fuego-adr new "Use PostgreSQL for persistence" --path docs/adr

# Edit the generated docs/adr/001-use-postgresql-for-persistence.adr.md, then:
fuego-adr serve docs/adr        # dev server with live reload at :8080
fuego-adr build docs/adr        # build the static site to build/
```

## Writing an ADR

An ADR is a Markdown file named `NNN-slug.adr.md` (e.g. `001-use-postgresql.adr.md`).
The leading number orders the decisions; `fuego-adr new` assigns it for you.

```markdown
---
title: Use PostgreSQL for persistence
status: accepted
date_proposed: 2026-01-10
date_accepted: 2026-01-15
author: fabio
approvers: [alice, bob]
tags: [database, infrastructure]
affects:
  - src/database/**
  - docker-compose.yaml
---

## Context

Why we needed to decide. The forces at play.

## Decision

What we chose, stated plainly.

## Consequences

What becomes easier or harder as a result.
```

### Frontmatter fields

| Field | Type | Notes |
|---|---|---|
| `title` | string | Decision title (required). |
| `status` | string | One of `tbd`, `proposed`, `accepted`, `deprecated`, `superseded`. |
| `date_proposed` | date | `YYYY-MM-DD`. |
| `date_accepted` | date | `YYYY-MM-DD`. |
| `deadline` | date | For `tbd` decisions — shown on the dashboard. |
| `author` | string or list | |
| `approvers` | list | |
| `tags` | list | Generates `/tags/{tag}/` pages. |
| `affects` | list | Glob patterns of files/areas this decision governs. |
| `supersedes` | list of ints | ADR numbers this one replaces. |
| `superseded_by` | list of ints | ADR numbers that replace this one. |

Single values may be written as a scalar (`author: fabio`) or a list; both are
accepted. Bare dates are normalized to `YYYY-MM-DD`.

### Sections

Each `## Heading` becomes a section. **`Context`, `Decision`, and `Consequences`**
get dedicated styling; any other heading (`Alternatives`, `Notes`, …) renders as
plain Markdown. For an `accepted` ADR, those three sections are **required** —
`build` and `validate` warn if one is missing.

### Supersession

When a decision replaces another, declare it on both sides:

```yaml
# in 004-replace-session-auth.adr.md
supersedes: [2]
# in 002-jwt-authentication.adr.md
superseded_by: [4]
```

`fuego-adr` validates that these references are mutual and point at ADRs that
exist; a mismatch fails the build.

## The generated site

| Page | URL | Contents |
|---|---|---|
| Dashboard | `/` | Decisions grouped by status (tbd / proposed / accepted / inactive). |
| Decision | `/decisions/{slug}/` | One page per ADR. |
| Timeline | `/timeline/` | All decisions newest-first, with supersession. |
| Affected files | `/affects/` | Each `affects` pattern → the decisions governing it. |
| Tags | `/tags/`, `/tags/{tag}/` | Tag index and per-tag listings. |

## CLI reference

```
fuego-adr new <title> [--path DIR]      Create a new ADR from a template
fuego-adr build [adr-path]              Build the static site
fuego-adr serve [adr-path]              Dev server with live reload
fuego-adr validate [adr-path]           Validate without building (CI gate)
fuego-adr list [adr-path]               List all ADRs
fuego-adr affected --files <f>...        Find ADRs whose `affects` match files
```

Common flags:

- `build`: `-o, --output DIR`, `--base-url PATH`, `--incremental`
- `serve`: `--base-url PATH`, `--port N`
- `list` / `affected`: `--json`, `-q, --quiet`

`[adr-path]` defaults to `.`.

### Examples

```bash
fuego-adr build docs/adr -o public --base-url /my-repo   # for a GitHub Pages subpath
fuego-adr validate docs/adr                              # CI: non-zero exit on any error
fuego-adr list docs/adr --json                           # machine-readable index
fuego-adr affected --files src/database/schema.sql -p docs/adr
```

## Configuration

Configuration is optional. Drop a `fuego-adr.yaml` next to (or above) your ADR
folder to set site-level metadata:

```yaml
site_name: "Engineering Decisions"
base_url: "/adr"          # deploy subpath; empty for root
output_path: "build"      # build output directory
```

Routes, taxonomies, the theme, and the parser all come from the ADR pack — you
only configure the site shell. CLI flags (`--output`, `--base-url`) override the
file.

## Deployment

Build with the base URL of your hosting path and publish the output directory.
For GitHub Pages under `https://user.github.io/repo/`:

```bash
fuego-adr build docs/adr --base-url /repo -o public
```

Then serve `public/` (e.g. via the Pages action or any static host).

## Using the ADR pack directly

fuego-adr is a thin CLI over `adr.Pack()`. In any Fuego project you can use the
pack directly — it brings its own parser, theme (including CSS/JS), routes,
taxonomies, and hooks:

```go
package main

import (
	"context"
	"log"

	"github.com/FabioSol/fuego-adr/adr"
	"github.com/FabioSol/fuego/engine"
)

func main() {
	eng := engine.New()
	eng.Use(adr.Pack())

	err := eng.Build(context.Background(), engine.BuildOptions{
		ContentDir: "docs/adr",
		OutputDir:  "build",
		SiteName:   "Engineering Decisions",
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

This is handy when you want ADRs as one section of a larger Fuego site, or to
combine the ADR pack with other packs and your own theme overrides (a file in
your `theme/renderers/decision.html` overrides the pack's).

## License

See the repository for license details.
