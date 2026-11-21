# Release Candidate Validation

Date: 2026-11-21. Product candidate: `26.1.0-rc.1`. Descriptor format: `0`.

## Passed

- Native Windows Go tests and vet, including initialization of a persistent deployment key, preservation of existing keys, and rejection of malformed existing keys.
- Linux race-instrumented tests for the application packages in an isolated Go build container.
- Frontend production builds, locale-key checks, desktop/mobile browser workflows, and all eight geographic-helper visibility combinations.
- Static experience tests for three preset examples, compact/padded names, decoded details, language/theme changes, project-subpath hosting, mobile width, and absence of API requests. Desktop and mobile screenshots were inspected.
- Clean multistage Docker build from an allowlisted source archive on the Linux trial host. An isolated release Compose project initialized its key, encoded/decoded a descriptor, recreated both services, and decoded the original name with the preserved key. Test containers and their disposable volume were removed.
- Runtime notices matched eight linked Go modules and Basecoat; Go, Tailwind, GeoNames and DB-IP notices were present.
- Go vulnerability scanning and production npm dependency auditing reported no known vulnerabilities on the test date. Gitleaks scanned the selected public source snapshot with redacted output and reported no leaks.
- GitHub workflow syntax was checked with actionlint. Windows amd64, Linux amd64 and Compose archives were generated with SHA-256 checksums.
- Twelve image-tag policy tests passed: exact candidates leave stable aliases unchanged; stable series/year/latest promotion uses numeric version ordering, all release pages, and the already-published image digest. Older-series and older-year patches preserve newer aliases. Failed release queries stop promotion.

## Publication checks

The public repository uses Git Flow with `develop`, `main` and `release/26.1.0`. Hosted [CI run 44873161145](https://github.com/Pipapoen/pipapoen-node-descriptor/actions/runs/44873161145) passed the Windows/Linux tests, Linux race and browser checks, security scans, clean container deployment and aggregate gate. Hosted verification identified and fixed explicit clipboard permissions for headless Chromium and the Go security baseline; CI and source containers now use Go 1.26.11. The branch protections require `CI gate`, and release tags are protected against updates and deletion.

The tag-triggered release workflow records the candidate's exact source commit and image digest. Verify anonymous image pulls and installation from the actual published deployment archive before promoting the candidate to stable.

The LAN MVP continues to use its existing deployment and key. Live GeoNames and DB-IP checks are recorded in MVP_VALIDATION.md. Real Authentik integration remains unconfigured; automated OIDC tests use a simulated provider.

## Shared browser experience

The revised Pages build uses the production frontend, styles and translations, with a WebAssembly build of the same Go codec and complete bundled geographic snapshot. The public key is 32 zero bytes. Dynamic external lookups are disabled. Native API-parity tests and browser ciphertext-vector checks verify consistent encoding/decoding; browser checks also cover custom fields and domains, imports/exports, clipboard reset, geographic filtering, language/theme switching, mobile layout and absence of API network traffic.
