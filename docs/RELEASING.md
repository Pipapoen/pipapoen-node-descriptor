# Release Operations

## Verification and branches

PND uses Git Flow with permanent `main` and `develop` branches. Require the `CI gate` status on pull requests into both branches. Restrict direct pushes and force pushes. CI covers Windows/Linux tests and builds, Linux race tests, browser flows, static demo behavior, dependency vulnerabilities, and a clean source-container deployment with persistent-key verification.

Create `release/26.1.0` from `develop` for the first release. Candidate versions use `26.1.0-rc.1`; stable versions use `26.1.0`. Update `package.json`, `internal/buildinfo/version.go`, source Compose images, and the deployment bundle defaults together. Run `python scripts/check-release.py` to check consistency. Product versions and descriptor versions are independent.

Descriptor format `0` retains the implemented pattern-assisted framing and projected resource dictionary. Freeze its field meanings, tier boundaries, provider codes and dictionary entries before the first stable release. New interpretations require a new descriptor version. UI suggestions and geographic names can evolve independently. Review format vectors and compatibility notes as part of release review.

## GitHub setup

Create the public repository as `Pipapoen/pipapoen-node-descriptor` using the authorized maintainer account. Use `develop` as the default contribution branch, configure branch protections, enable GitHub Pages with the GitHub Actions source, and create a `release` environment with maintainer approval where available. Restrict that environment to release tags. Keep workflow permissions read-only by default; the publishing job requests package/content writes separately.

The repository must permit its Actions token to publish its GHCR package. Set the package visibility to public and verify anonymous pulls before announcing a release. Third-party Actions are pinned to reviewed commit SHAs; Dependabot proposes monthly updates.

## Publish

1. Review the candidate source, complete CI, check licenses and scan the exact public file set for credentials and local artifacts.
2. Tag a reviewed commit on `release/26.1.0` as `v26.1.0-rc.1`. Push that tag to run the release workflow. Candidate tags must point to commits in the matching release branch.
3. The workflow reuses CI, validates tag/version agreement, builds Linux amd64 and Windows amd64 archives, and publishes a Linux amd64 GHCR image with provenance and an SBOM. It creates a prerelease with deployment archives, checksums, and the image digest. Candidate images retain their full version tag. Stable releases also advance eligible series, year and `latest` aliases after their release assets are published.
4. Download the candidate Compose archive into an empty directory and verify anonymous image pull, first launch, restart, round trip, and upgrade with the existing key volume. Review the demo and README from their published locations.
5. Set the stable product version, validate it, merge the release branch into `main` and back into `develop`, then tag the reviewed `main` commit `v26.1.0`. Stable tags must point to commits contained in `main`.

Treat published version tags and assets as immutable. Correct a published release with a new patch version. A failed publication can leave an image or draft release; inspect those artifacts before retrying. Publish release notes only after all required assets are available.

## Image tags

| Tag | Update scope |
| --- | --- |
| `26.2.3` | Immutable exact release |
| `26.2` | Highest published stable patch in this release series |
| `26` | Highest published stable version in this year |
| `latest` | Highest published stable version across all years |
| `26.3.0-rc.1` | Exact candidate; candidates leave stable aliases unchanged |

Aliases reference the already-built image digest, preserving its provenance and SBOM. Promotion compares numeric versions across all pages of published GitHub releases. Drafts and prereleases are excluded. For example, publishing `26.2.4` after `26.3.0` advances only `26.2`. Publishing an older-year fix can advance that year's aliases while preserving the global `latest`.

Release publication jobs share one concurrency group. A failed release query stops alias promotion. Failed promotion can be retried with the reviewed `RELEASE_VERSION`, `IMAGE_DIGEST` and authenticated GitHub/Docker clients using `python scripts/image_tags.py`; exact release images are retained. Keep the release history available for version comparison. Publish one release at a time and inspect any canceled pending workflow before continuing.

Deployment bundles pin the exact image version. Operators can choose a moving alias through `PND_IMAGE`; `docker compose pull` followed by `docker compose up -d` installs its current image.

## Static experience

GitHub Pages serves a separate Vite build with three preset examples, compact/fixed output, decoded fields, English/Traditional Chinese, and light/dark/system themes. The examples are generated by the real codec using a public synthetic key. The page calls no PND API and accepts no private node data. Runtime application assets do not include this demo key or example fixtures.

Regenerate examples with `go run ./scripts/demo-fixtures`, build with `pnpm run build:demo`, and test with `pnpm run test:demo`. CI checks that the generated fixtures match the committed files. The Pages workflow deploys validated static files from `main`.

## Release evidence

Record the reviewed source commit, CI run, package checksums, container digest, platform coverage, dependency scan date, and deployment test outcome. Live Authentik verification remains an operator integration check until a real client is available. GeoNames and DB-IP live integration were verified in the MVP environment; CI uses local fixtures and optional-service checks to avoid account and upstream availability dependencies.
