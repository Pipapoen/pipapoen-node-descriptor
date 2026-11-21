# MVP Validation

Date: 2026-11-20. Build: `26.1.0-mvp.1`. Descriptor format: experimental `0`.

## Completed checks

- TypeScript checks and Vite production build passed. Initial built frontend assets were approximately 26 kB JavaScript and 10 kB CSS before compression.
- English and Traditional Chinese locale keys and values passed validation.
- Go tests passed on Windows. Encoding tests cover 512 optional-field masks, low/high boundary values, compact/padded output, normalized IPs, independent Python-derived packing vectors, deterministic output, and changed keys.
- A three-second decode fuzz campaign completed 66,839 inputs without failure.
- API tests covered malformed/unknown JSON fields, cross-origin rejection, public operation, and unconfigured reference-service behavior.
- A simulated OIDC provider exercised discovery, PKCE, callback state, token verification, nonce rejection, session issuance, and protected API access. Cookie tests rejected modified and expired sessions. Partial OIDC configuration failed validation.
- Linux amd64 codec and server test binaries passed on the trial host. The deployed application uses the cross-compiled Linux binary, embedded frontend, host CA bundle, and collected third-party licenses in a scratch image.
- Playwright with the installed Edge browser passed desktop and mobile workflows against both the local service and the directly reachable LAN deployment. Checks included provider changes preserving selected RAM, generation, 63-character padding, decode/edit, identity preservation across locale changes, themes, guide access, and horizontal-overflow detection.
- Docker Compose started the dedicated PND service with its independently generated remote key. Health checks passed and the service was reachable directly on the LAN.

## Observed resources

The deployed image measured 8,542,899 bytes (about 8.54 MB). Post-start and post-browser-check idle samples reported 3.973 and 6.094 MiB memory with 0.00% CPU, against a 128 MiB container limit. These are observations from this host, not peak-memory guarantees.

A low-rate LAN trial completed 50 encode/decode pairs with exact recovery. Median pair latency was 26.06 ms and the sampled 95th percentile was 40.81 ms. Each pair included two HTTP requests. This trial validates ordinary interactive use; it does not establish concurrency or saturation limits.

## Verification boundaries

- A real Authentik client has not been configured. OIDC has been validated with a simulated provider; the trial deployment is intentionally public.
- Live GeoNames search and ID lookup passed on November 21 after account configuration. DB-IP GeoIP verification is recorded below. Direct GeoNames ID input remains available independently.
- The trial image was built remotely from reviewed precompiled artifacts. The full source multistage Dockerfile is supplied for normal builds; a clean source-container build has not yet been exercised in this environment.
- Windows build/tests and Linux execution passed. Race-instrumented tests, security audit, sustained load, other CPU architectures, and formal release automation remain future checks.
- The historical evaluation documents retain comparison results. Current behavior and the single-key contract are defined in TECHNICAL.md and DEPLOYMENT.md.

## Trial operation

The service is managed as Compose project `pnd-mvp` in `/opt/pnd-mvp` on the selected trial host. Its encryption key remains in the deployment's secret directory. Updating the key and restarting changes the active key; exported node data is the recovery source for generating replacement names. DNS records are managed separately.

GitHub CI, GHCR publication, and formal release work follow the MVP feedback stage.

## Interface revision

The November 20 interface revision adopts Tabler CSS components and a compact sidebar/form layout. The generated stylesheet is approximately 78 kB (13.8 kB gzip), including the application overrides, after unused CSS removal. Tabler JavaScript is excluded.

Browser regression checks additionally verified read-only Node ID, an initially identity-only descriptor, clearing optional values, preservation of meaningful numeric zero values, new-ID generation, imported ID retention, and actual background-color changes between dark and light modes. The format, active key, and existing descriptor compatibility remain unchanged.

## Modern navigation and geographic tools revision

The following revision replaces Tabler with MIT-licensed Basecoat components and a neutral top navigation. It removes the duplicate product heading and focus/active underlines from the brand control. The secondary brand name is configured by `PND_BRAND_NAME`; a local trial used `Example Lab` and verified the value in the API and rendered interface.

Country/city browsing, online place search, and IP lookup now share a helper toolbar beneath GeoNames ID. The bundled GeoNames snapshot contains 252 country/region entries and 34,248 major cities/capitals, compressed to 737,943 bytes of country-specific city files. Tests verify every country's count, page limits, distinct pages, city filtering, invalid inputs, and Taipei ID 1668341. Browser checks verify explicit application, preservation of GeoNames ID when switching countries, and IP-tool placement. The dataset retains CC BY 4.0 attribution and its coverage limits.

The revised frontend build is approximately 61 kB CSS (9.5 kB gzip) and 33 kB JavaScript (12.1 kB gzip). Component-library JavaScript is excluded. Desktop, narrow-screen, locale, theme, geographic-tool, and codec flows passed against both the local build and the directly reachable remote deployment. Linux API and geography tests passed. The revised image measured 9,388,297 bytes; one post-browser-check sample reported 8.867 MiB memory and 0.00% CPU. These measurements include the embedded geographic snapshot and do not establish peak load limits.

## City localization

The localization pilot covered Taiwan, Japan, the United States, Germany, and France before the full snapshot was transformed. Assertions verified stable IDs and localized names for Taipei, Tokyo, New York, Berlin, and Paris, plus explicit source-name fallback. The full transform found usable Chinese names for 11,296 of 34,248 city records. Country-specific compressed city files total 1,555,060 bytes after adding localized names, administrative-area names, and search aliases. Internal search aliases are stripped from API responses.

Go tests cover Chinese and English queries reaching the same IDs and all-country pagination/count consistency. Browser tests check Traditional Chinese display and English search in the same locale. OpenCC runs only during data construction and adds no runtime service dependency. Upstream name coverage remains incomplete; these checks do not claim that every city has a verified Chinese translation.

## Documentation navigation revision

The in-app Guide page and its navigation entry were removed. Browser regression checks assert that both are absent while encoding, decoding, language/theme changes, and geographic helpers continue working. README now contains a short introduction and illustrated encoding, location-selection, and decoding instructions. Its three screenshots were captured from the actual interface using synthetic example data, and all local documentation/image links were verified. Technical and deployment details remain in dedicated documents; contextual field help remains in the interface.

The following interface revision increases light-theme border contrast and integrates country/city search inputs into expanding selection lists. Browser checks cover mouse selection, keyboard Enter/Escape, retained selections, localized filters, and dropdown layering. README screenshots now consistently use United States / New York example data (GeoNames ID 5128581).

The November 21 search revision hides city sections and historical/abandoned/destroyed settlements by default, adds a details toggle, and ranks exact city names ahead of administrative-area matches. Tests confirm New York City's exact-name result ranks first and ordinary administrative seats remain eligible. Separate lookup tests recover a default-hidden place from the full reference index. Codec tests recover arbitrary valid numeric GeoNames IDs without reference-table membership. Browser checks verify decoded country/city labels, preservation of a missing numeric ID, and a failed/missing name lookup leaving decoding usable. The descriptor format and key remain unchanged.

The helper-configuration revision tests all eight combinations of country/city, place-search, and IP-lookup flags. Go tests verify actual endpoint gates, disabled database handling, manual geographic encoding, and independent local ID lookup. A browser test mocks effective metadata for all eight combinations to verify visibility and keyboard tab navigation; normal browser flows separately exercise the deployed default configuration. Invalid boolean settings fail validation. Go's govulncheck v1.10.0 reported no known vulnerabilities for Windows and Linux amd64 source analysis on 2026-11-21; the scope and remaining OIDC verification limits are documented in AUTHENTICATION.md.

The short-session revision retains stateless signed cookies and adds `PND_SESSION_TTL`, defaulting to 30 minutes. Deterministic-clock tests cover the exact expiry boundary, future/missing issue times, shorter settings applied to existing cookies, larger settings preserving the original expiry, and independent server instances accepting the same valid cookie. The OIDC flow test verifies a configured 15-minute lifetime in both the signed payload and cookie, and confirms that API activity does not issue a replacement cookie. Invalid duration settings are rejected. Live Authentik verification remains pending; the LAN trial continues in public mode.

## Account-free GeoIP

The one-shot updater downloaded the official DB-IP Lite City 2026-11 MMDB, verified gzip/MMDB structure, and recorded raw-file SHA-256 `5f64491f77c2c78b56cd83a88db4ea6c4666874322da38805498bbb754ccc735`. The database file is 126,873,350 bytes and is kept outside the application image. A second invocation reported `current` and reused the same-month file.

Unit tests cover successful installation, same-month reuse, invalid/truncated gzip, invalid database content, failed initial download, preservation of old data after failed refresh, and the writer lock. Geographic candidate tests cover exact country/city matching and empty/unknown results. The real local database returned usable city results for two IPv4 addresses and one IPv6 address. Browser tests verify explicit candidate selection, unchanged network fields, private-address rejection, and visible DB-IP attribution. These are point-in-time lookup checks rather than proof of location accuracy.

The updater is a finite Compose job with a separate data volume; runtime queries remain local. Periodic scheduling is not enabled. The operator can run the documented refresh command and restart PND to load a new mapped file. A missing or failed update without a usable database blocks initialization instead of installing an invalid file.

Remote Linux updater/geography/API tests passed, followed by an actual account-free download into the Compose data volume. The next initialization reported `current` and exited successfully. Direct LAN browser tests passed for both address families and explicit candidate selection. One observed runtime sample used 16.05 MiB memory and 0.00% CPU; the updater container had exited with code 0. The final updater also checks an existing file against its recorded hash before allowing it as a failed-refresh fallback; a corrupted-cache regression test passed. OIDC remains disabled on the trial deployment.

## Live GeoNames verification

On November 21, the LAN deployment enabled GeoNames using an operator-supplied account in its private deployment configuration. Search for New York returned eight results including ID 5128581. Direct online lookup of that ID returned New York, country US. The metadata endpoint reported both place search and GeoIP enabled, and the health endpoint returned OK. Public example configuration retains an empty GeoNames username.
