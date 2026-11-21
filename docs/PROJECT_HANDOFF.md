# Pipapoen Node Descriptor (PND)
## Project Background and Requirements

- Original date: 2026-11-18
- Document version: v0.1; protocol details remain under evaluation.
- English translation and requirements addendum: 2026-11-19.
- Original positioning: an internal Pipapoen node-information encoder/decoder and reversible data format.
- Status: historical requirements background. The current MVP behavior is described in docs/TECHNICAL.md and docs/DEPLOYMENT.md.

This document preserves the original requirements and tradeoffs. Section 11 records subsequent user decisions and takes precedence where the original internal-only deployment assumptions differ. The concise architecture is documented in [the technical specification](TECHNICAL.md). Tentative protocol details remain open to technical evaluation.

## 1. Background

Pipapoen is a community for technical exchange and resource and service sharing. Nodes may be rented VPS instances, dedicated servers, owned physical machines, member-provided devices, or other infrastructure. They should not all be called VPS instances or named after their current services, such as LXD, proxy, or storage.

The idea originated with a reinstalled LXD host being prepared for Authentik OIDC. An initial naming proposal was `node-<provider>-<region>-<sequence>.pipapoen.org`. Discussion identified several limitations:

1. Nodes have no natural sequence; centralized sequence allocation requires occupancy state.
2. Locations need identifiers that are unambiguous across countries; city names and custom abbreviations are insufficient.
3. Names can carry more descriptive information without being directly human-readable.
4. Encoded values should be recoverable without loss, while outsiders should not easily recover location, IP addresses, or configuration by guessing a conversion rule.
5. The calculator should not require a central database tracking whether each Node ID has been used.

The resulting scope is a reversible Node Descriptor format with optional confidentiality, independent optional fields, and a dynamic capacity budget, carried in a single DNS label, plus an encoder/decoder application.

## 2. Product Scope and Boundaries

### 2.1 In scope

- Let users enter or select node information and choose which fields to include.
- Show field bit costs, structural overhead, remaining capacity, and expected hostname length in real time, using a capacity-budget interaction model.
- Generate a valid single DNS label, originally illustrated as `n<encoded-payload>.pipapoen.org`. The fixed one-character prefix is `n`; the earlier candidate `p` was rejected.
- Decode Node ID, present fields, exact values, and resource tiers.
- Use GeoNames ID as the sole standard geographic identifier. Manual entry, place search, and IP lookup are alternative ways to obtain it.
- Avoid a central Node Registry or sequence-occupancy lookup. Server-side keys, read-only reference data, and necessary caches are permitted.

### 2.2 Out of scope

- A complete node rental, resource scheduling, monitoring, or LXD management system.
- Automatic creation of Cloudflare DNS records, TLS certificates, or OIDC clients.
- Treating a Descriptor as proof of ownership, truthful resources, administrator permissions, or a trusted credential.
- Highly dynamic runtime values such as CPU/RAM utilization, free disk space, or online status.
- Sacrificing compactness to recover exact hardware specifications; CPU, RAM, and storage may use approximate tiers.
- A stateful allocation service solely to avoid duplicate sequence numbers.

## 3. User Flow and Interface

The earlier P0/P1/P2/P3 profile hierarchy has been discarded. The interface uses a mandatory core, fully independent optional fields, and a dynamic bit pool. User-facing categories are:

| Category | Contents |
| --- | --- |
| Node identity | Generate, import, or retain a random Node ID; explain that regeneration changes identity |
| Geography | Enter a GeoNames ID, search places, or use IP geolocation as assistance |
| Hosting source | Select or enter a provider or other hosting source; identifier source remains unresolved |
| Resources | CPU architecture, virtualization capability, CPU capacity, RAM, and storage tiers |
| Network | Independently optional IPv4 and IPv6; both or neither may be included |
| Capacity and results | Used/remaining capacity, field costs, DNS length, copy output, and decoding |

When capacity is insufficient, users must explicitly choose the tradeoff. Never silently discard data or reduce promised precision. Show the bit cost and length effect of each option.

### Geographic input methods

1. Direct GeoNames ID: look up and display the place name, country, administrative area, feature class/code, and other confirmation details.
2. Place search: support country and city-name searches. Default to GeoNames class `P` (Populated Place), avoiding mixed city, administrative-area, and airport results for names such as Taipei. Show feature code, administrative area, and ID. An advanced search may be offered.
3. Public IP assistance: use GeoIP, with MaxMind GeoLite2 City as a candidate, to suggest a city GeoNames ID and show uncertainty. Users may accept or modify it. Do not invent an ID when city-level data is absent. Show differences between IPv4 and IPv6 results instead of arbitrarily choosing one.

All methods populate the same proposed 32-bit GeoNames ID field. User-confirmed locations take precedence over GeoIP estimates. A GeoNames ID identifies a geographic database record; it does not guarantee one record per metropolitan area. City/metropolitan consolidation rules remain undecided and must not be assumed implicitly.

## 4. Core Data Model

### 4.1 Node identity

- Generate random Node IDs using the OS CSPRNG; the tentative size is 80 bits.
- Do not depend on central sequence numbers or IP uniqueness: multiple nodes may share a NAT public address.
- Uniqueness is probabilistic, not an absolute mathematical guarantee of no collisions.
- Descriptors with different detail levels for the same node should decode to the same Node ID.
- Users or nodes must retain their original Node ID to preserve identity across re-encoding. Absence of central state does not eliminate the need to retain individual identity data.

### 4.2 Optional attributes

Fields are independently optional. Lifecycle groupings are only interface categories and hints about when changes require re-encoding.

- Placement: GeoNames ID and Provider ID.
- Resources: CPU architecture, virtualization capability, CPU capacity tier, RAM tier, and storage tier, each independently optional.
- Network: full IPv4 (32 bits) and full IPv6 (128 bits), independently optional.
- Runtime: excluded from the Descriptor.

Resource tiers are intentionally lossy summaries. Mapping an exact RAM value to an 8-16 GiB tier allows recovery only of that tier. Node ID, GeoNames ID, Provider ID, IPv4, and IPv6 must be fully reversible once included.

### 4.3 Provider is not synonymous with ASN

A short integer Provider ID is a candidate, but a reliable globally unique source is not established. An ASN identifies an autonomous network, not necessarily a VPS reseller, hosting provider, or individual member. A Pipapoen-maintained static Provider Registry is an option; its vocabulary maintenance is distinct from a stateful per-node registry. If no trustworthy Provider ID is available, leave the field unselected.

## 5. DNS and Bit Budget

### 5.1 DNS form

- Use a single subdomain label, originally `n<encoded-payload>.pipapoen.org`, rather than an additional level such as `...infra.pipapoen.org`. The original deployment preference was compatibility with first-level wildcard coverage from Cloudflare Universal SSL.
- The fixed prefix `n` is mandatory. The complete first label is limited to 63 ASCII octets.
- The candidate alphabet is lowercase `0-9a-z` (Base36). Use a unique canonical representation consistent with case-insensitive DNS.
- After the prefix, at most 62 Base36 characters remain, with approximately 320.5 bits of theoretical capacity.

### 5.2 Tentative bit allocation

| Field | Bits | Proposed meaning |
| --- | ---: | --- |
| Header | 8 | Candidate: 3-bit format version and 5 top-level presence flags |
| Random Node ID | 80 | Mandatory, CSPRNG-generated |
| GeoNames ID | 32 | Optional, complete integer |
| Provider ID | 16 | Optional, identifier rules unresolved |
| Resources presence bitmap | 5 | Present only when any resource field is included |
| CPU architecture | 2 | Optional approximate enum |
| Virtualization capability | 3 | Optional state/capability |
| CPU capacity | 4 | Optional tier |
| RAM capacity | 4 | Optional tier |
| Storage capacity | 3 | Optional tier |
| IPv4 | 32 | Optional complete address |
| IPv6 | 128 | Optional complete address |
| Total with all fields | 317 | Excludes a separate authentication tag, key version, and extensions |

Resource values total 16 bits, plus the 5-bit resource bitmap. The five top-level flags may represent Geo, Provider, Resources Group, IPv4, and IPv6. When the resource group is absent, its bitmap consumes no space. This header interpretation and layout are proposals requiring verification.

Ideal packing of 317 bits requires `ceil(317 / log2(36)) = 62` Base36 characters, reaching 63 with the prefix. There is no meaningful room for additional overhead. Variable length, leading zeros, field representation, key versions, authentication, and exceptional values all affect feasibility.

The bit pool must calculate actual serializable/encrypted length for the selected fields. Information-theoretic estimates alone do not demonstrate a valid format.

## 6. Encoding, Decoding, and Confidentiality

### 6.1 Conceptual pipeline

```text
User-selected fields and values
  -> normalization, resource tier mapping, and GeoNames ID confirmation
  -> version + Node ID + presence bitmap + optional fields
  -> reversible canonical binary packing and DNS-safe representation
  -> keyed reversible encryption (FF1 FPE is a candidate, not confirmed)
  -> n<encoded-payload>.pipapoen.org
```

Decoding reverses the pipeline and uses GeoNames/Provider reference data for readable names. The format must be reversible without a per-node encoding lookup table. Identical inputs with the same format and key must behave predictably. Different format versions must be recognized correctly or explicitly rejected.

### 6.2 Encryption requirements

Base36, lossless compression, seeded character substitution, and custom dictionaries do not provide trustworthy confidentiality. Server-side format-preserving encryption (FF1) is being considered to reversibly transform DNS-safe strings without increasing their length, so outsiders cannot easily recover cities, resources, or origin IP addresses.

FF1 evaluation covers variable-length payloads, all field combinations, canonical encoding, available libraries, and maintenance:

- FF1 provides confidentiality, not integrity or source authentication. Successful decryption does not establish trust or identity.
- Keys require secure backups. Rotation and key-version identification affect whether old hostnames remain decodable.
- Fixed tweaks and deterministic encryption reveal repeated inputs; variable lengths may reveal approximate configurations.
- Keys must never be sent to browsers, including in authenticated deployments. Authorized encoder/decoder users may make guessing queries.
- If confidentiality, integrity, every optional field, and 63 characters cannot coexist, explain capacity/functionality tradeoffs. Do not substitute homemade obfuscation for encryption.

Descriptors are names and descriptions. Never grant permissions based on decoded values or treat them as proof of location or hardware specifications.

## 7. Reference Data and Service Boundaries

- GeoNames supplies standard `geonameId` values, place search, and ID lookup. Populated-place filtering assists selection; duplicate names, administrative areas, representative cities, and metropolitan membership may still require confirmation.
- MaxMind GeoLite2 City is a candidate for suggesting city GeoNames IDs from user-supplied public IPv4/IPv6 addresses. It is not authoritative and may lack city results or precise facility locations.
- A versioned static Provider Registry may identify hosting sources. It maintains a vocabulary, not registered nodes.
- Authentik OIDC was the original internal-access integration. Integrating this calculator is separate from integrating LXD with Authentik. See Section 11 for configurable SSO and public mode.
- The Node Descriptor Service performs encoding, decoding, and reference queries without saving every Descriptor or sequence allocation. Persistent keys and necessary read-only databases/caches are compatible with this goal.

DNS may be hosted with Cloudflare, but PND must not depend on Cloudflare as its only runtime. Do not infer origin IPs from Cloudflare proxy DNS addresses. Node providers explicitly supply IP addresses; GeoIP only assists geographic input.

## 8. Identity and Hostname Lifecycle

- Node ID represents persistent node identity.
- A Descriptor represents a selected set of attributes; changing data or field selection may change its string.
- Using a Descriptor as an LXD management hostname means changes to IP, RAM, location, or provider may affect TLS, DNS, Caddy, and Authentik OIDC redirect URIs.
- The interface must distinguish Node ID from the complete Descriptor/hostname. Whether frequently changing Descriptors should serve as permanent management DNS names remains unresolved.
- Encryption is not authorization. An internal decoder still requires access controls and consideration of abuse.

## 9. Established Direction and Open Decisions

### Established original direction

- Working name: Pipapoen Node Descriptor.
- A calculator service, not a central node-management or occupancy database.
- One DNS label with the `n` prefix.
- Mandatory core, independent optional fields, and a live bit pool; no user-facing profile hierarchy.
- IPv4 and IPv6 may both be included or independently omitted.
- GeoNames ID is the sole stored geographic identifier, with direct entry, search, and IP assistance.
- Users control field inclusion; overflow must not cause silent data loss.
- Resources use approximate tiers, not exact hardware reconstruction.
- Reversible encoding is fundamental; server-held standard encryption is a candidate confidentiality layer.

### Open technical decisions

1. Complete bit layout: validate the proposed 317 bits and overhead; consider Node ID/Provider widths, bitmap changes, or explicit restrictions on extreme combinations.
2. Resource tier boundaries and reserved values such as `unknown`/`other`; decide whether virtualization requires flags instead of an enum.
3. A trustworthy Provider ID source and sensible behavior when omitted.
4. City/metropolitan semantics: promise only populated places or define a traceable mapping; do not claim GeoNames inherently provides one-to-one metropolitan identifiers.
5. Canonical representation: variable lengths, leading zeros, integer boundaries, version identification, unambiguous parsing, and behavior before/after encryption.
6. Encryption and integrity: FF1 suitability, tamper detection, key versions, backward compatibility, and query abuse within capacity limits.
7. Long-term hostname stability: distinguish permanent management endpoints from detailed Descriptors if appropriate.
8. MVP scope: usable encoding/decoding and geographic assistance without excessive design complexity.

The final protocol depends on the capacity and security tradeoffs identified above. Bit allocation and encryption selection remain open.

## 10. Original Reference Sources

These links are preserved from the original draft; their current status has not been revalidated as part of this translation.

- DNS labels and hostnames: RFC 1035 <https://www.rfc-editor.org/rfc/rfc1035>; RFC 1123 <https://www.rfc-editor.org/rfc/rfc1123>
- GeoNames API: <https://www.geonames.org/export/web-services.html>
- GeoNames search and filtering: <https://www.geonames.org/export/geonames-search.html>
- GeoNames feature codes: <https://www.geonames.org/export/codes.html>
- MaxMind City database fields: <https://dev.maxmind.com/geoip/docs/databases/city-and-country/>
- MaxMind web-service responses, including `city.geoname_id`: <https://dev.maxmind.com/geoip/docs/web-services/responses/>
- NIST SP 800-38G (FF1): <https://csrc.nist.gov/pubs/sp/800/38/g/upd1/final>
- NIST SP 800-38G Rev.1 second public draft (identified as a draft in the original document): <https://csrc.nist.gov/pubs/sp/800/38/g/r1/2pd>

## 11. Confirmed Requirements Addendum (2026-11-19)

These subsequent user decisions extend and, where necessary, supersede the original draft:

- Public distribution is planned under the Pipapoen GitHub organization with the MIT license.
- Make deployment configurable for other operators, including domain, SSO, keys, and external data services. Pipapoen-specific examples must not become hardcoded deployment requirements.
- The backend uses Go; the frontend uses Vite with vanilla TypeScript and CSS. Established, maintained packages provide standard functionality.
- Review direct/transitive dependency licenses and separate data/asset terms for compatibility with the intended MIT distribution. Preserve required third-party notices; project licensing does not replace dependency licenses.
- Target Windows and Linux. Docker Compose is the primary deployment method; the authorized remote Linux environment supplies Docker validation.
- Provide a web UI with optional OIDC SSO. Configured SSO requires authentication for application operations and APIs. Without SSO configuration, the calculator is public. Invalid enabled SSO configuration must fail closed rather than silently enabling public access.
- Do not introduce a local user database, roles, or permission-management system in the first version. Necessary short-lived authentication sessions are distinct from a user-management database.
- Keep secrets out of source control, browser bundles, logs, and published images. Public decoding cannot provide confidentiality from visitors authorized to call that decoder.
- Establish CI testing and a formal release process using GitHub Actions and GHCR container images, without requiring Docker Hub.
- Use calendar versions `YY.<release-sequence>.<patch-sequence>`, such as `26.1.0`, `26.1.1`, and `26.2.0`. Keep product and Descriptor format versions independent.
- Require low resource consumption. Prefer a single application service and avoid unnecessary databases, Redis, or background processes. Bound caches, request sizes, and concurrency. Evaluate geographic data storage/memory costs and measure idle/load consumption before setting acceptance budgets. Do not trade away security validation for lower overhead.
- Support UI language switching. The initial supported locale list and locale-selection policy remain to be decided.
- Support `light`, `dark`, and `system` themes; `system` follows the operating-system preference. Preference persistence and defaults remain implementation decisions.
- Contextual help and a dedicated usage page provide detailed guidance. Field labels, selection states, and layout communicate the primary actions.
- The README presents the project introduction and informative badges, including AGAIN AI Disclosure. The technical specification describes architecture and behavior for users and maintainers.
