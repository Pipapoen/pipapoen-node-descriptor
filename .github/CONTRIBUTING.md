# Contributing

PND uses Go and Vite with TypeScript. See [local development](../docs/DEPLOYMENT.md#local-development) for setup and [translations](../docs/TRANSLATING.md) for language contributions.

## Repository layout

| Directory | Contents |
| --- | --- |
| `cmd/`, `internal/` | Go entry points, application packages and embedded data |
| `web/` | Shared frontend and Vite, TypeScript and PostCSS configuration |
| `demo/` | Browser demo entry point and worker transport |
| `deploy/` | Release Compose template, source environment example and optional overlays |
| `docs/` | Technical documentation, project background and screenshots |
| `scripts/` | Build, validation and packaging tools |
| `third_party/` | Dependency and dataset notices |
| `build/` | Local binaries, release archives, static site, test results and deployment bundles |

Run development commands from the repository root. `compose.yaml` and `Dockerfile` provide the source-build entry points. Release archives place Compose files together for standalone deployment.

## Branches and commits

Use Git Flow. `main` holds release-ready history and `develop` integrates development. Create `feature/*`, `fix/*`, or `docs/*` branches from `develop`, and submit a pull request back to `develop`. Use Conventional Commits, such as `feat: add a location filter`.

Create `release/<version>` from `develop` for release stabilization. Merge the validated release into `main` and back into `develop`. Create `hotfix/*` from `main` for production fixes and integrate the fix into both lines. Delete merged topic branches; retain `main` and `develop`.

## Verification

Run `pnpm run build`, `pnpm run check:locales`, and `go test ./...` for application changes. CI also runs Go vet, Linux race tests, browser workflows, the static demo, and a clean container deployment. The required aggregate status is `CI gate`.

Changes to the descriptor format require explicit compatibility review and stable test vectors. Existing format-version meanings and decoding dictionaries remain immutable. Dependency updates include license notices and vulnerability checks.

Keep credentials, deployment settings, generated binaries, downloaded GeoIP databases, and local agent files out of commits. Review the selected files before committing. See [release operations](../docs/RELEASING.md) for publishing.
