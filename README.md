# WailsDoc

WailsDoc generates API documentation directly from Wails v2 Go controllers.

It discovers exported controller methods, parameters, return values, DTOs,
GoDoc, JSON tags, source locations, type links, and backlinks without a manual
manifest or annotations.

## Installation

```bash
go install github.com/AmadoMuerte/wailsdoc/cmd/wailsdoc@latest
```

## Quick Start

```bash
cd my-wails-app
wailsdoc init --package ./internal/transport/wails --ui vitepress --name "My App"
npm install --prefix docs/site
wailsdoc generate
wailsdoc serve
```

For CI, commit generated output and run:

```bash
wailsdoc check
```

## Configuration

```yaml
version: 1
project:
  name: My Wails App
  title: My Wails App API
scan:
  packages:
    - ./internal/transport/wails
output:
  schema: docs/generated/wails-api.json
  markdown: docs/generated/api
  inventory: docs/wails-api-inventory.json
ui:
  provider: vitepress
  directory: docs/site
  packageManager: auto
wails:
  bindings: frontend/src/wailsjs/go/backend
validation:
  forbiddenFields:
    - password
    - sessionkey
```

CLI flags override configuration where supported.

## Package Manager

`ui.packageManager` selects the package manager used by `wailsdoc serve` and
`wailsdoc build`. Accepted values are `auto` (default), `bun`, `npm`, `pnpm`,
and `yarn`; any other value is a configuration error.

With `auto`, WailsDoc inspects the UI directory and then the repository root for
lockfiles and picks the matching manager:

| Marker | Manager |
| --- | --- |
| `bun.lock`, `bun.lockb`, `bunfig.toml` | `bun` |
| `pnpm-lock.yaml`, `pnpm-workspace.yaml` | `pnpm` |
| `yarn.lock` | `yarn` |
| `package-lock.json` | `npm` |

When no marker is found it falls back to `npm`, so existing projects keep their
current behavior. Set the field explicitly to skip detection.

For a Bun project:

```yaml
ui:
  provider: vitepress
  directory: docs/site
  packageManager: bun
```

```bash
bun install --cwd docs/site
wailsdoc generate
wailsdoc serve
```

## Markdown Only

Use `ui.provider: none`. `wailsdoc generate` then requires only Go and writes
the JSON schema and portable Markdown documentation. Node is only required by
`wailsdoc serve` and `wailsdoc build` for VitePress projects.

## UI

The first release supports the standard VitePress theme with automatic
controller and type navigation, local search, syntax highlighting, responsive
layout, and dark mode. Markdown-only output is always available.

WailsDoc targets Wails v2 controller conventions. Wails v3 is not supported.

More projects: [amadomuerte.ru](https://amadomuerte.ru).
