# Deployment and Development

PND runs as one Go process with embedded frontend assets. New names use frozen format `1`; existing format-0 names remain decodable with their original key. Export node data and back up the key before upgrades. Applications released before rc.4 support format 0; upgrade consumers before generating format-1 names.

## Release bundle

The release archive `pnd-<version>-compose.zip` contains `compose.yaml`, `.env.example`, optional overlays, and documentation. Copy `.env.example` to `.env`, edit the browser origin and hostname settings, and run `docker compose up -d`. Compose reads `.env` automatically in this separate deployment directory and pulls the pinned GHCR image. Source builds use `deployment.env` as described below.

A one-shot `init-key` service creates a 32-byte encryption key in the `keys` named volume and exits. It validates and preserves an existing key. The PND container mounts the volume read-only. The default project name is `pnd`, producing volume `pnd_keys`; custom project names change the volume name. Preserve the project name and back up the key volume. An invalid existing key stops initialization for operator inspection. Deleting the volume changes the identity of a subsequent deployment and prevents correct recovery of existing names.

Change `PND_IMAGE` in `.env` to the desired version, then run `docker compose pull` and `docker compose up -d`. Include the same optional overlay files when updating an instance that uses them. The packaged overlays share the configured image version. For packaged commands below, `.env` replaces the source checkout's `--env-file deployment.env` argument, and overlay filenames omit the `deploy/overlays/` prefix.

Existing file-key deployments can continue using the source Compose configuration. Migrating to the release bundle requires preserving the original key and optional database volume; a fresh release volume creates a separate installation.

## Docker Compose

Build the application with Docker Compose v2 or newer:

```sh
docker compose build
mkdir -p secrets
chmod 700 secrets
docker run --rm --user "$(id -u):$(id -g)" -v "$PWD/secrets:/secrets" pnd:26.1.0-rc.4 -generate-key /secrets/pnd-key
chmod 644 secrets/pnd-key
cp deploy/source.env.example deployment.env
docker compose --env-file deployment.env up -d
```

The secret directory limits host access while the mounted key file is readable by container user 65532. An existing key file is preserved by the generator. Set `PND_PUBLIC_URL` to the actual browser origin in `deployment.env`, such as an internal HTTP IP and port or a public HTTPS origin. The default port binds to all host interfaces for direct LAN access. Docker secrets mount the key at `/run/secrets/pnd-key`.

On Windows, create `secrets` in PowerShell and run the image with an absolute bind-mount path. Windows bind-mount access follows Docker Desktop and host filesystem permissions:

```powershell
docker compose build
New-Item -ItemType Directory -Force secrets
docker run --rm --mount "type=bind,source=$((Resolve-Path secrets).Path),target=/secrets" pnd:26.1.0-rc.4 -generate-key /secrets/pnd-key
Copy-Item deploy/source.env.example deployment.env
docker compose --env-file deployment.env up -d
```

Health is available at `/healthz`. The default Compose budget is 128 MiB, half a CPU, and 64 processes. `GOMEMLIMIT=96MiB` gives the Go runtime a soft memory target. Filesystem writes and Linux capabilities are restricted in the application container.

Set `PND_BRAND_NAME` in `deployment.env` to customize the name beside PND. The default is `Pipapoen`; an explicitly empty value hides the secondary name. The value is plain display text, up to 64 characters, and is independent of the output domain and descriptor format. Recreate the service after changing environment configuration.

The repository uses a `.env/` directory for local tooling. Compose configuration uses `deployment.env` explicitly with `--env-file`.

## Generated hostname defaults

Set `PND_PREFIX` and `PND_DOMAIN` in `deployment.env` to initialize the hostname fields:

```dotenv
PND_PREFIX=x
PND_DOMAIN=example.org
```

These are the public defaults for both native execution and Docker Compose. The prefix accepts one lowercase letter (`a` through `z`). Set the domain to a DNS suffix you operate. Users can adjust both fields in the interface. `PND_PUBLIC_URL` separately identifies the browser origin of the application. Recreate the container after changing deployment settings.

## Single encryption key

`PND_KEY_FILE` must identify a file containing 64 hexadecimal characters, representing one 32-byte AES key. The application fails startup when the file is missing or malformed. The key is loaded once at startup and remains server-side.

Replace the key file and restart or recreate the service to change keys. Existing descriptors require the original key for correct recovery. The service retains one active key, and the label carries only its format version. Exported node data provides the inputs for regenerating names after key changes. Updating DNS and downstream URLs is an operator action.

## Optional OIDC sign-in

Create an OIDC client with callback `<PND_PUBLIC_URL>/auth/callback`. Configure `PND_OIDC_ISSUER` and `PND_OIDC_CLIENT_ID` in `deployment.env`, and place the client secret in `secrets/oidc-client-secret` with container-readable permissions.

```sh
docker compose --env-file deployment.env -f compose.yaml -f deploy/overlays/compose.oidc.yaml up -d
```

OIDC discovery, Authorization Code with PKCE, state, nonce, ID-token verification, and signed HttpOnly session cookies are supported. Public HTTPS origins produce Secure cookies. Session identity is not written to a user database. Logout clears the PND session; the identity provider's session is managed by that provider.

Set `PND_SESSION_TTL=30m` in `deployment.env` to control the absolute session lifetime. The default is 30 minutes; examples include `15m` and `1h`. Whole-second durations from `1m` through `24h` are accepted. Missing configuration uses the default; invalid values prevent startup. Recreate the container after changing its environment.

The signed cookie records issue and expiry times, and the server validates both the signed expiry and the currently configured maximum age. API activity does not renew a session. Shortening the setting limits existing timestamped sessions by their original issue time; increasing it does not extend an existing signed expiry. Older cookies without an issue time require a new login after this update. Browser cookie expiry matches the issued lifetime.

Sessions remain stateless across instances sharing the signing key and configuration. No server-side session records or back-channel logout endpoint are introduced. Expiry requires a new OIDC flow; whether the user must enter credentials again is controlled by the identity provider's own session policy. IdP account changes are not pushed immediately to an already issued PND session.


Incomplete OIDC configuration fails startup. Without any OIDC configuration, the calculator is public. Application APIs require a session when OIDC is enabled. HTTPS issuers are required by default; `PND_ALLOW_HTTP_OIDC=true` is an explicit development-only option for an HTTP test issuer. A live Authentik connection requires the operator's issuer and client configuration.

## Optional geographic services

Each helper has an independent boolean environment setting, defaulting to `true`:

| Setting | Helper |
| --- | --- |
| `PND_COUNTRY_CITY_ENABLED` | Bundled country/city selector |
| `PND_PLACE_SEARCH_ENABLED` | Live GeoNames place search and its helper endpoint |
| `PND_IP_LOOKUP_ENABLED` | GeoIP address lookup |

Set a value to `false` in `deployment.env` and recreate the service to disable it. Disabled helper APIs return `404` with `feature_disabled`, and the corresponding UI tab is hidden. Place search also requires a GeoNames username; IP lookup also requires a valid database. Unconfigured helpers are hidden. Disabling IP lookup skips opening its configured database. With no available helpers, the entire helper toolbar is hidden.

Manual GeoNames ID entry, encoding/decoding, and local ID-to-name resolution remain available independently. Disabling place search also disables its live ID-lookup endpoint, so optional external name fallback is unavailable while local decoded names continue working.

The GeoNames ID helper includes a bundled country/major-city selector requiring no external account. Its GeoNames cities15000 snapshot contains cities over 15,000 population or capitals, rather than every settlement. City lists load by country with filtering and 100-record pagination. Selecting a country preserves the existing GeoNames ID; applying a city updates that field. The snapshot source, coverage, hashes, and CC BY 4.0 attribution are retained in `internal/geography/data/manifest.json` and `third_party/licenses/geonames_NOTICE.txt`.

Country and city controls open a list with an integrated search box. Country filtering accepts the displayed localized name, English name, or two-letter country code. City filtering searches the full selected-country snapshot by city name, ASCII name, or administrative area, then paginates the matching results. Filtering preserves the selected country and the existing GeoNames ID. Mouse selection and Enter close the list; Escape dismisses it. Additional city pages can be loaded within the dropdown.

City and administrative-area names now include English and Traditional Chinese where supplied by GeoNames alternateNamesV2. The builder prefers zh-TW, zh-Hant, and zh-HK entries, then converts generic Chinese through OpenCC s2twp at build time. Historic/ended names are excluded from display candidates. Search accepts retained English and Chinese aliases regardless of interface language. The snapshot includes Chinese names for 11,296 of 34,248 cities; other records display English or the source name. Geographic IDs remain unchanged.

The embedded snapshot is static and currently dated 2026-11-20. The running service performs no scheduled upstream synchronization. Updating it requires downloading reviewed GeoNames source files, regenerating the compact data with `scripts/build-geography.py`, updating the snapshot date and attribution, testing, and rebuilding/redeploying the application. Optional account-based place-name search queries the GeoNames service live and is separate from this bundled selector.

Localization uses `uv run --script scripts/localize-geography.py --phase pilot`, followed by `--phase full` after the pilot passes. The script pins its build-only OpenCC dependency and filters the large upstream alias archive down to the embedded geographic records. A release-time or monthly reviewed snapshot refresh is the recommended future workflow; it is not currently scheduled. A failed refresh should leave the last verified snapshot in service.

- Set `PND_GEONAMES_USERNAME` to enable GeoNames place search and ID lookup through its HTTPS API. Direct numeric ID entry is available without a configured account. Search results show administrative context and feature codes.
- Obtain an appropriately licensed GeoIP2/GeoLite2 City database and place it at `data/GeoLite2-City.mmdb`. Start Compose with `-f compose.yaml -f deploy/overlays/compose.geoip.yaml` to mount it read-only. Restart after updating the database. Public IPv4 and IPv6 suggestions are displayed separately and applied only by user selection.
- Optional lookups fail visibly when data is unavailable; selected form values remain intact. GeoNames and MaxMind require their applicable attribution and data-use terms.

Country/city selection, online place search, and IP lookup are grouped beneath GeoNames ID. The IP helper accepts a separate address or the currently selected network addresses. Running a lookup does not add a network field to the descriptor.

## Account-free GeoIP download

Use the optional one-shot DB-IP Lite City updater to initialize a named data volume:

```sh
docker compose --env-file deployment.env -f compose.yaml -f deploy/overlays/compose.geoip-auto.yaml up -d --build
```

Compose builds the application image, runs `geoip-update`, and starts PND after successful initialization. The updater exits after completion; it is not a resident service. This option requires no account. The IP helper displays DB-IP attribution, and the separate data volume retains the downloaded database and a release/hash sidecar. It is an alternative to the manually mounted `compose.geoip.yaml` configuration; select one GeoIP overlay.

The updater discovers the published monthly MMDB link on the official HTTPS source page, enforces compressed/uncompressed size limits, verifies gzip integrity and MMDB structure/type, and records SHA-256. The hash records the downloaded file; it is not a publisher signature. Valid same-month data is reused. On a failed refresh, a usable existing database remains available. With no usable existing database, initialization fails visibly. Separate files are replaced individually after validation, with a writer lock to prevent concurrent updates.

Refresh explicitly and restart PND to load the new mapped file:

```sh
docker compose --env-file deployment.env -f compose.yaml -f deploy/overlays/compose.geoip-auto.yaml run --rm geoip-update
docker compose --env-file deployment.env -f compose.yaml -f deploy/overlays/compose.geoip-auto.yaml restart pnd
```

There is no background schedule. A deployment can invoke this update command during its normal monthly maintenance. The application mounts the data volume read-only; only the one-shot updater writes it. For native Windows updates, stop readers before replacement. If an updater is forcibly killed, inspect its state before removing a stale `.lock` file.

DB-IP Lite city records supply names and approximate locations rather than a city GeoNames ID. PND offers exact country/name matches from its GeoNames snapshot, with administrative context and explicit user confirmation. Multiple candidates remain choices. If no reliable match exists in the snapshot, the result remains descriptive and no identifier is invented. A manually supplied MaxMind database can continue to provide its native city GeoNames ID.

## Local development

Published native archives contain the application and license/documentation files. Extract the archive, create a `secrets` directory, and run `pnd -generate-key secrets/pnd-key` once. Start `pnd` from that directory and open `http://localhost:8080`. Use `./pnd` on Linux and `./pnd.exe` in PowerShell; the same `PND_*` environment settings apply. Back up the key before upgrades. Linux extraction tools should preserve executable permissions; use `chmod +x pnd` if needed.

Use Go 1.26.11 or newer, Node.js 24, and the pinned pnpm version. From the repository root:

```sh
pnpm install --frozen-lockfile
pnpm run build
go test ./...
go build -o build/pnd ./cmd/pnd
mkdir -p secrets
./build/pnd -generate-key secrets/pnd-key
./build/pnd
```

On Windows, use `go build -o build/pnd.exe ./cmd/pnd`, `New-Item -ItemType Directory -Force secrets`, and `./build/pnd.exe`. The default development key path is `secrets/pnd-key`. The frontend build must precede Go builds because assets are embedded.

`pnpm dev` serves the Vite frontend and proxies API/auth requests to the Go service at `127.0.0.1:8080`. `pnpm run check:locales` validates translation keys. `pnpm run test:browser` uses an installed Microsoft Edge and tests the service at `PND_TEST_URL` or `http://127.0.0.1:8080`.

## Trial scope

The MVP includes encoding, decoding, capacity previews, independent fields, provider-aware recommendations, JSON import/export, English/Traditional Chinese, light/dark/system themes, and contextual help. Illustrated operation steps are provided in the README. It generates names; operators create DNS records separately. GitHub CI, GHCR images, release archives, and the static experience are described in RELEASING.md.
