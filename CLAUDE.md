# CLAUDE.md — fuego-adr Contributor Guide

## What is fuego-adr?

fuego-adr is a **domain-specific static site generator for Architecture
Decision Records**, built on the [Fuego](https://github.com/FabioSol/fuego)
meta-engine. It turns `*.adr.md` files into a documentation site with a status
dashboard, timeline, affected-files index, tag pages, and a page per decision,
with section-completeness and supersession validation.

It is structured as a **Fuego format pack** (`adr.Pack()`) plus a thin CLI. The
pack is the product; the CLI is a zero-config convenience wrapper over it.

## How it uses Fuego

fuego-adr does **not** fork or modify Fuego. Everything works through Fuego
v0.3's public extension points:

| Fuego extension point | fuego-adr usage |
|---|---|
| `core.Parser` + `core.FilenameParser` | `adr.New()` parses `*.adr.md` (filename glob) |
| `core.Pack` (`eng.Use`) | `adr.Pack()` bundles parser + theme + config + hooks |
| `Pack.Theme fs.FS` | embedded `adr/theme/` (templates + `static/` CSS/JS) |
| `Pack.ConfigDefaults` | `adr/config-defaults.yaml` (route + tags taxonomy), deep-merged |
| `core.AfterParseHook` | number extraction, supersession + section validation |
| `core.IndexHook` | dashboard / timeline / affects-index virtual pages |
| `engine.Build/Serve/Validate` | the programmatic build API the CLI drives |
| `core.Node{Raw: true}` | each ADR section is pre-rendered HTML |

If something here feels limiting, the fix usually belongs in Fuego's pack API,
not in a workaround here (see AD-6).

## Architecture Decisions

### AD-1: fuego-adr is a format pack, not an engine wrapper

**Decision:** All domain logic lives in the importable `adr` package, assembled
into a single `core.Pack` returned by `adr.Pack()`. The CLI just calls
`eng.Use(adr.Pack())` and the programmatic build API.

**Why:** A pack is the v0.3 unit of reuse. It lets ADRs be one section of a
larger Fuego site, lets users override the theme per-file, and keeps fuego-adr
out of the business of orchestrating the pipeline. A vanilla project needs one
line: `eng.Use(adr.Pack())`.

### AD-2: The parser owns section structure; the engine stays generic

**Decision:** `ADRParser.Parse` splits the Markdown body on `## ` headings and
emits one `core.Node` per section, with `Type` = the lowercased heading
(`context`, `decision`, `consequences`, …) and `Content` = rendered HTML marked
`Raw: true`. Content before the first heading becomes a `preamble` node.

**Why:** Per-type nodes let the theme style `Context` / `Decision` /
`Consequences` differently via `theme/renderers/{type}.html`, while the engine
never needs to understand ADR structure. Rendering Markdown to HTML in the
parser (goldmark) and flagging `Raw` means the default renderer passes it
through unescaped.

### AD-3: Virtual pages are generated in an Index hook

**Decision:** The dashboard (`/`), timeline (`/timeline/`), and affects-index
(`/affects/`) are virtual `*core.Page`s appended in an `IndexHook`, not a
`BeforeRender` hook.

**Why:** Index hooks run during INDEX, after ROUTE has resolved real-page URLs
(so summaries can link to `/decisions/{slug}/`) and before the collision
re-check — so these virtual pages are validated against real and taxonomy pages
like any other. (In the pre-v0.3 version this ran in `BeforeRender` and bypassed
collision detection.)

### AD-4: Enrichment and validation run in an AfterParse hook

**Decision:** `AfterParseHook` extracts the ADR number from the filename into
the envelope, validates bidirectional supersession consistency (a hard error),
checks required sections for `accepted` ADRs (a warning), and defaults the
layout to `adr`.

**Why:** This must happen before ROUTE/INDEX so the number and layout are set
when URLs are resolved and summaries are built. Supersession is a correctness
property worth failing the build over; section completeness is advisory.

### AD-5: The pack ships its own theme and static assets

**Decision:** The entire theme — `base.html`, layouts, renderers, and a
`static/` directory with the compiled Tailwind CSS and JS — is embedded in the
pack via `//go:embed theme` and exposed as `Pack.Theme`.

**Why:** A pack should render a complete site with no files in the consumer's
project. Fuego v0.3 copies a pack theme's `static/` subtree to the output root
during STATIC, so the CSS/JS travel with the pack. A user's own
`theme/renderers/*.html` or `public/*` still override the pack's.

### AD-6: Config comes from defaults + the engine's option overrides, never a temp file

**Decision:** ADR routes and the tags taxonomy live in
`adr/config-defaults.yaml` (`Pack.ConfigDefaults`, deep-merged under the site
config). Site-level settings (content dir, output dir, site name, base URL) are
passed through `engine.BuildOptions`. fuego-adr never writes a temporary
`config.yaml`.

**Why:** Earlier versions synthesized a config string and a temp theme dir and
shelled into `eng.Run(--config ...)`. v0.3's `ConfigDefaults` + programmatic
`engine.Build/Serve/Validate` remove all of that. `fuego-adr.yaml` is therefore
tiny — only `site_name`, `base_url`, `output_path`.

## Project Structure

```
fuego-adr/
  main.go                     CLI entry point → internal/cli.Execute()
  adr/                        the format pack (importable)
    adr.go                    Pack() — assembles parser + theme + config + hooks
    parser.go                 *.adr.md parser, section split, frontmatter normalize
    hooks.go                  AfterParse (enrich/validate) + Index (virtual pages)
    config-defaults.yaml      routes + tags taxonomy (Pack.ConfigDefaults)
    theme/                    embedded theme
      base.html, layouts/, renderers/, static/
  internal/
    cli/                      Cobra commands: build, serve, validate, new, list, affected
    config/                   optional fuego-adr.yaml (site_name, base_url, output_path)
  testdata/sample/            sample ADRs for manual builds and tests
```

`internal/cli` depends on `adr` and `internal/config`; `adr` depends only on
Fuego (`core`, `engine`) and goldmark.

## Build flow

`fuego-adr build docs/adr` does:

1. `loadConfig` reads `fuego-adr.yaml` (or defaults).
2. `engine.New()`; `eng.Use(adr.Pack())`.
3. `eng.Build(ctx, BuildOptions{ContentDir: adrPath, OutputDir, SiteName, BaseURL, Incremental})`.

Inside Fuego, the pack contributes the parser (so `*.adr.md` is discovered as
content), the theme, the route/taxonomy config defaults, and the two hooks. The
options layer sets the dirs and site metadata on top. `serve` and `validate`
follow the same assembly via `eng.Serve` / `eng.Validate`.

## Common Tasks

### Change how a section renders
Edit `adr/theme/renderers/{section}.html` (e.g. `decision.html`). The node's
`.Content` is the section's rendered HTML.

### Add a new validated section
Add the heading to `requiredSections` in `adr/parser.go` and create a renderer
for it. Authors add `## NewSection` to accepted ADRs.

### Change ADR routing or add a taxonomy
Edit `adr/config-defaults.yaml`. The user's config still wins via deep-merge.

### Add a generated overview page
Append a virtual `*core.Page` in `IndexHook` (set its `URL`, `Layout`, and
`Envelope`) and add a matching layout in `adr/theme/layouts/`. It will be
collision-checked automatically.

### Regenerate the stylesheet
The theme CSS is compiled Tailwind (`tailwind.css` source → `adr/theme/static/style.css`).
Rebuild it with the project's Tailwind tooling before committing.

## Testing

- `go test ./...` — unit tests for the parser and the pack.
- `adr/pack_test.go` is the key regression: it builds a site with **only**
  `eng.Use(adr.Pack())` (no CLI) and asserts the decision page, dashboard,
  timeline, affects-index, tag pages, and pack static assets all render. Keep
  it passing — it is the "vanilla pack" contract.
- `go run . build testdata/sample -o /tmp/out` for a manual end-to-end build.

## Dependency note

`go.mod` may carry a `replace` directive pointing at a local Fuego checkout
while consuming unreleased engine features. Before publishing, pin a tagged
Fuego release and remove the replace.
