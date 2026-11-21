# Third-Party Notices

PND source is licensed under MIT. Dependencies retain their respective licenses.

The following license and notice files accompany the application runtime:

- [github.com/coreos/go-oidc/v3](third_party/licenses/github.com_coreos_go-oidc_v3_LICENSE)
- [github.com/coreos/go-oidc/v3](third_party/licenses/github.com_coreos_go-oidc_v3_NOTICE)
- [github.com/go-jose/go-jose/v4](third_party/licenses/github.com_go-jose_go-jose_v4_LICENSE)
- [github.com/joshuajorel/scrambler](third_party/licenses/github.com_joshuajorel_scrambler_LICENSE)
- [github.com/oschwald/geoip2-golang/v2](third_party/licenses/github.com_oschwald_geoip2-golang_v2_LICENSE)
- [github.com/oschwald/maxminddb-golang/v2](third_party/licenses/github.com_oschwald_maxminddb-golang_v2_LICENSE)
- [golang.org/x/oauth2](third_party/licenses/golang.org_x_oauth2_LICENSE)
- [golang.org/x/sys](third_party/licenses/golang.org_x_sys_LICENSE)
- [golang.org/x/time](third_party/licenses/golang.org_x_time_LICENSE)
- [Go runtime and standard library](third_party/licenses/go_LICENSE)

The frontend includes selected [Basecoat](https://basecoatui.com) components by Ronan Berder under the [MIT license](third_party/licenses/basecoat_LICENSE). Unused selectors are removed at build time. Component-library JavaScript is not shipped. Frontend build tools and browser testing dependencies are pinned in `pnpm-lock.yaml` and retain their own package licenses.

The precompiled component stylesheet retains the [Tailwind CSS MIT license](third_party/licenses/tailwindcss_LICENSE).

The frozen resource dictionary was evaluated against public DigitalOcean, Akamai/Linode, Vultr, and AWS Lightsail plan specifications collected on 2026-11-20. See the source links and normalization notes in [the catalog evaluation](docs/RESOURCE_DICTIONARY_EVALUATION.md). Catalog suggestions are snapshots.

The bundled GeoNames country, administrative-area, and cities15000 snapshot uses CC BY 4.0; see [attribution and transformations](third_party/licenses/geonames_NOTICE.txt). Optional live GeoNames lookup uses the same source attribution. Optional GeoLite2 data is supplied separately by the operator under MaxMind terms; no GeoIP database is bundled.

The optional account-free updater downloads DB-IP Lite City under CC BY 4.0. See [DB-IP attribution and processing](third_party/licenses/dbip_NOTICE.txt). The database is stored in a separate data volume and retains its source attribution.

The expanded provider recommendation catalog records selected factual specifications from official provider pages and public APIs. Source URLs, check dates and response hashes are recorded in `internal/catalog/source-plans.json`; transformation and coverage details are in [Catalog](docs/CATALOG.md). PND-authored self-managed and flexible-shape examples are labeled explicitly. Vendor marketing text and logos are excluded from the catalog.
