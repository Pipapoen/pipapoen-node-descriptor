# Provider and Recommendation Catalog

The November 21, 2026 snapshot contains 40 provider choices, 587 recommendations and 62 product series. Recommendations include reviewed provider specifications and explicitly labeled PND reference configurations. They assist descriptor entry; regional availability and configurable options are checked through each row's source link.

## Provider IDs

Provider IDs belong to the PND registry. Existing IDs retain their meanings, and new providers receive unused IDs. Names and aliases can be updated when a provider changes branding. One descriptor contains one optional Provider ID.

| ID | Provider |
| --- | --- |
| 0 | Self-managed |
| 1 | DigitalOcean |
| 2 | Akamai Cloud / Linode |
| 3 | Vultr |
| 4 | AWS |
| 5 | Hetzner |
| 6 | OVHcloud |
| 7 | Scaleway |
| 8 | UpCloud |
| 9 | Oracle Cloud |
| 10 | Google Cloud |
| 11 | Microsoft Azure |
| 12 | Alibaba Cloud |
| 13 | Tencent Cloud |
| 14 | IONOS |
| 15 | Contabo |
| 16 | netcup |
| 17 | OneProvider |
| 18 | DediRock |
| 19 | RackNerd |
| 20 | BuyVM |
| 21 | HostHatch |
| 22 | GreenCloud |
| 23 | Crunchbits |
| 24 | CloudCone |
| 25 | Hostinger |
| 26 | Kamatera |
| 27 | Leaseweb |
| 28 | Servers.com |
| 29 | DataPacket |
| 30 | V.PS |
| 31 | BandwagonHost |
| 32 | DMIT |
| 33 | LightNode |
| 34 | Servarica |
| 35 | HostDare |
| 36 | Clouvider |
| 37 | InterServer |
| 38 | ScalaHosting |
| 65535 | Other |

`internal/catalog/providers.json` stores names, aliases and official websites. The selector searches those names, aliases and numeric IDs. Imported IDs outside the current registry remain usable and display numerically. Directory membership is independent of plan coverage: providers with no reviewed recommendations remain selectable and offer a direct website link for manual resource entry.

## Recommendations

`internal/catalog/source-plans.json` stores reviewed capacity facts, source URLs, check dates and source-response SHA-256 values. The raw web responses are retained only as local research artifacts. Each record names its product series, CPU allocation, memory, optional included disk, architecture and node type. PND reference configurations and Oracle flexible-shape examples are labeled separately from fixed provider products.

The builder converts explicit byte units to GiB before selecting the existing resource interval. Decimal MB/GB/TB and binary MiB/GiB/TiB retain distinct meanings. Linode and Vultr API memory counters use their 1024-based sizing convention; Linode disk counters use the same convention. Scaleway reports memory in bytes. Values explicitly labeled GB in reviewed page tables are retained as GB. The UI shows the source capacity text, while the descriptor stores its enclosing resource interval.

Oracle E5 examples convert one x86 OCPU to two vCPUs; A1 examples use one vCPU per OCPU. Flexible examples select documented CPU/memory combinations and leave separately configured storage open. Optional local disks, volume limits and maximum disk counts are excluded from included-storage values.

Applying a recommendation sets its provider, CPU, memory and documented architecture/node type. Unspecified architecture is preserved. Included storage is applied when present; separately configured storage preserves the current selection. Node identity, location and network addresses are retained. Searching and filtering preserve the descriptor until a recommendation is applied. Resource menus continue to offer every valid tier, with matching catalog tiers grouped first.

The snapshot covers selected general-purpose, compute, memory, shared/dedicated-vCPU and Arm offerings. It retains distinct product names even when several map to the same encoded tiers. Hetzner cost-optimized entries describe published specifications with variable stock. OVHcloud's source currently labels its reviewed range VPS 2027. Prices and live availability are outside this snapshot.

## Update and contribute

1. Review an official specification page or public provider API. Record its URL, check date, capacity units and relevant product family. Check pagination when using an API.
2. Add or update facts in `source-plans.json`. Add new providers to `providers.json` using unused IDs. Preserve prior ID meanings. Mark authored configurations as `example` or `configuration`.
3. Run `python scripts/build-catalog.py`. The generated files are `plans.json` and `catalog-info.json`; the runtime reads only these compact advisory data and the provider registry.
4. Run `python scripts/build-catalog.py --check`, `python -m unittest discover -s scripts -p "test_*.py"`, `go test ./internal/catalog ./internal/codec`, and the catalog browser checks.
5. Review changed mappings and source facts before deploying the new snapshot. A rebuild publishes the update; the running application performs no catalog polling.

Formats 0 and 1 continue to use the frozen `dictionary.json`. Provider menus and recommendations evolve independently of that dictionary. Its canonical SHA-256 is `744533032e5282766388956d37298cd41f2be0c6ba9e3a999cc897d03db0bed3`; compatibility tests verify it alongside existing codec vectors.

OneProvider recommendations cover published OneCloud general-purpose, high-performance and storage dimensions. Rows differing only in traffic allowances are combined. DediRock recommendations retain the published Los Angeles/New York KVM and Storage plan names. Architectures are omitted when the referenced product page does not state them.
