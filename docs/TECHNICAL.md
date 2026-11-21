# PND Technical Specification

Status: release candidate `26.1.0-rc.5`, descriptor format `1` frozen for the first stable release, with format-0 decoding compatibility.

## Architecture

PND is one Go service serving an embedded Vite/TypeScript frontend and JSON API from the same origin. Users retain node data through JSON import/export. Optional OIDC provides sign-in through an external identity provider; sessions use signed, HttpOnly cookies with a configurable absolute lifetime (30 minutes by default). Omitted OIDC configuration enables public access. Incomplete configuration or failed discovery prevents startup.

The application targets Windows and Linux. Docker Compose is the primary deployment method. The runtime container runs as user 65532 with a read-only filesystem and a mounted encryption key. Reference queries have timeouts and bounded responses. APIs enforce a 16 KiB request-body limit and a shared 30 requests/second limit with a burst of 60.

The three geographic helpers have independent deployment flags: `PND_COUNTRY_CITY_ENABLED`, `PND_PLACE_SEARCH_ENABLED`, and `PND_IP_LOOKUP_ENABLED`. Their defaults are true; live helpers additionally require their reference configuration. Metadata advertises effective availability, unavailable tabs are hidden, and explicitly disabled endpoints return 404. Manual GeoNames ID values and local decoded-name resolution remain independent. See [deployment configuration](DEPLOYMENT.md) and [authentication maintenance](AUTHENTICATION.md).

## Descriptor format

Labels use `<prefix>1<ciphertext>` under a configurable domain. The prefix is one lowercase letter; `1` selects the frozen codec. Existing format-0 names remain decodable. Each complete label is limited to 63 ASCII characters.

| Field | Bits | Semantics |
| --- | ---: | --- |
| Node ID | 80 | Mandatory random identity retained across changes to the same node |
| GeoNames ID | 32 maximum | Optional positive geographic identifier |
| Provider ID | 16 maximum | Optional registry identifier |
| Architecture | 2 | x86-64, ARM64, RISC-V 64, other |
| Node type | 3 | Physical, VM, system container, application container, reserved, other |
| CPU | 4 | Assigned logical CPU/quota interval |
| RAM | 4 | Assigned memory interval in GiB |
| Storage | 3 | Selected local-storage capacity interval in GiB |
| IPv4 | 32 | Optional complete address |
| IPv6 | 128 | Optional complete address without a zone identifier |

Optional fields are independent. Identifiers and addresses recover exactly; resources recover to their selected intervals. Code zero is distinct from an absent field. Exact bit order, interval boundaries, canonical rules and compatibility are specified in [Descriptor Format 1](FORMAT_1.md).

Packing uses a leading-one sentinel, a pattern-assisted selection header, fixed/varint geographic and provider integers, independent architecture/node-type values, and a projected CPU/RAM/storage dictionary with direct-field fallback. The embedded dictionary freezes 73 resource triples. Canonical-form and structural validation run during decoding. [The codec evaluation](CODEC_COMPARISON.md) explains the comparison.

The integer becomes lowercase Base36. Compact mode encrypts its shortest representation; padded mode first left-pads to 61 characters. The decoder accepts both automatically, decrypts before removing padding, and restores bit length from the sentinel. Cosmetic prefix and domain changes preserve recovery.

## Encryption and keys

Format `1` uses FF1, radix 36, AES-256, and tweak `PND:format-1:ff1:aes256` over the entire payload. The FF1 dependency is pinned in `go.mod` and retains its Apache-2.0 license. PND source uses MIT.

The service loads one 32-byte key from `PND_KEY_FILE` at startup. Missing or malformed keys prevent startup. Replacing the file and restarting changes the key. Names carry a format version and no key-generation identifier. Operators export node data to regenerate names after key changes and update DNS and downstream URLs.

FF1 preserves length and equality for identical inputs under the same parameters. Padding conceals original length. Descriptors provide confidentiality without a cryptographic authenticity tag. Modified ciphertext or a wrong key may produce plausible fields. Service sessions determine access; public-mode callers can decode names supported by that deployment.

## API

| Route | Purpose |
| --- | --- |
| `GET /healthz` | Health and build version |
| `GET /api/session` | Sign-in requirement and session status |
| `GET /api/meta` | Intervals, providers, suggestions, defaults, lookup availability |
| `POST /api/node-id` | Generate a random Node ID |
| `POST /api/preview` | Validate fields and calculate bit/character costs |
| `POST /api/encode` | Return descriptor, label, hostname, and capacity |
| `POST /api/decode` | Recover data from a label, hostname, or HTTP(S) URL |
| `GET /api/places` | Optional GeoNames search (`q`) or ID lookup (`id`) |
| `GET /api/geoip` | Optional public-IP city suggestion (`ip`) |
| `GET /api/countries` | Bundled country/region list and city counts |
| `GET /api/cities` | Country-filtered major cities (`country`, optional `q`, `offset`), 100 per page |
| `GET /api/location` | Resolve a numeric GeoNames ID (`id`) against the complete reference snapshot |
| `/auth/login`, `/auth/callback`, `POST /auth/logout` | OIDC and local logout |

Preview/encode receive `descriptor` plus optional `prefix`, `domain`, and `mode`. Descriptor properties are `nodeId`, `geo`, `provider`, `architecture`, `nodeType`, `cpu`, `ram`, `storage`, `ipv4`, and `ipv6`. Tier values are numeric codes; omitted properties remain absent. Decode receives `input`. Stable error codes are translated by the frontend.

## Interface and reference data

English and Traditional Chinese are included, with light/dark/system themes. The interface uses selected MIT-licensed Basecoat components, with unused CSS removed during the build. A compact top navigation replaces the original sidebar. `PND_BRAND_NAME` customizes the secondary brand label independently of the domain. Browser language/theme choices persist independently. Node ID is read-only in the form; generation, JSON import, and decoded-node reuse supply its value. Optional fields use an empty value to represent absence, including empty dropdown options. Numeric code zero remains a selected value. Provider-aware recommendations reorder dropdown groups while preserving values. Applying a suggestion explicitly changes CPU/RAM/storage. The README provides illustrated user instructions; contextual help remains beside relevant controls.

Each descriptor contains at most one Provider ID. Its 16-bit domain contains 65,536 codes; `0` denotes self-managed and `65535` denotes other, leaving 65,534 potential provider codes. The review registry includes 38 hosting providers alongside those two categories, with searchable names and aliases. UI read-only handling prevents accidental ID edits; imported IDs are validated as 80-bit values rather than attested as service-generated identities.

GeoNames ID entry works without external configuration. A GeoNames account enables search and ID lookup. An operator-supplied GeoIP City database enables separate IPv4/IPv6 suggestions. Users confirm suggested locations. Reference data does not determine node identity.

An optional one-shot DB-IP Lite City updater supports account-free GeoIP deployment. It downloads the official monthly MMDB, validates it before replacement, and retains usable old data after failed refreshes. The runtime reads the database locally; no IP addresses are sent to an external lookup service. DB-IP city names produce user-confirmed GeoNames candidates from exact country/name matches. Native IDs in compatible MaxMind databases remain supported. See the automatic-download and manual-mount alternatives in DEPLOYMENT.md.

The built-in country/city helper uses an attributed GeoNames cities15000 snapshot: 252 country/region entries and 34,248 major-city/capital records at the November 20 snapshot. With English/Traditional Chinese alternate names and search aliases, approximately 1.56 MB of country-grouped compressed city JSON is embedded. Chinese names are available for 11,296 records; others retain English/source names. Each query reads only one country's data, searches names and aliases, removes internal search aliases from the response, and paginates at 100 records. The selector, online search, and IP helper share a toolbar beneath GeoNames ID. Changing country or lookup text preserves existing encoded values until explicit application. The snapshot is static; reviewed data refresh and redeployment remain separate from live account-based search.

The default city picker hides PPLX, PPLH, PPLQ, and PPLW records. Its details toggle sends `details=true` to include them. Ordinary populated places, capitals, and administrative seats remain available. Search ranking places exact display-name matches first, followed by prefixes, contained names, and administrative-area/alias matches; population order breaks ties. This changes presentation only and does not merge or rewrite IDs.

After decoding, the frontend separately resolves the retained GeoNames ID to localized country and city names. The ID-to-country reference index and lookup use the entire snapshot, including records hidden by default in the picker. An optional configured GeoNames live lookup can supplement a missing local record. Missing or failed name resolution leaves the numeric ID and the successful descriptor decode intact. Reference labels and search filters do not alter codec versions, ciphertext, or field values.

## Validation and release path

Tests cover field masks, boundaries, independent packing vectors, malformed input, deterministic output, key changes, cookies, OIDC state/nonce/PKCE, and API protection. Browser tests cover generation, padding, decoding, retained values, languages, themes, and narrow layouts. Linux execution is verified on the trial deployment.

GitHub workflows verify the application, publish versioned GHCR images and release archives, and deploy the static experience. The release Compose bundle initializes a persistent key volume on first use. Product versions follow `YY.release.patch`; descriptor versions evolve independently. See [release operations](RELEASING.md), [MVP validation](MVP_VALIDATION.md), [deployment](DEPLOYMENT.md), and [translation contributions](TRANSLATING.md).

## Provider catalog

The review snapshot adds searchable provider names/aliases and product-series recommendations with explicit source units, provenance and optional included storage. See [catalog maintenance and ID allocation](CATALOG.md). The shared format-0/1 decoding dictionary remains frozen.

OIDC browser entry redirects unauthenticated users directly to the identity provider. Local logout displays a confirmation with a sign-in action. Failed callbacks display a retry page; API session expiry starts a new OIDC flow.
