# Resource Dictionary Evaluation from Provider Catalogs

Observed: 2026-11-20. Status: measured format candidate; resource tiers remain provisional.

## Sources and selection

Official public sources supplied 127 plan rows. DigitalOcean's rendered default tables supplied 31 Basic, CPU-Optimized, General Purpose, Memory-Optimized, and Storage-Optimized plans. Linode's public API supplied 48 current rows in nanode, standard, dedicated, and highmem classes with no successor. Vultr's public API supplied 37 paid vc2, vhf, and vhp rows. AWS Lightsail's pricing page supplied 11 Linux general-purpose bundles with public IPv4. Both API responses were checked for complete pagination. Hetzner's fetched page did not expose a usable static plan table and was excluded.

These are available catalog configurations, with equal row weights in the main evaluation. They do not establish sales volume, installed-base frequency, or community-node prevalence. GPU plans, arbitrary physical servers, custom volumes, containers, and ARM configurations are outside this sample. Architecture/environment codes are fixed to the experiment's x86/VM interpretation for these rows; final resource semantics remain undecided.

## Capacity normalization

The source values are retained as nominal provider capacity numbers. DigitalOcean specifies GiB/MiB. Linode and Vultr API memory values are divided by 1024; Linode disk values are also divided by 1024. Vultr and Lightsail disk GB values retain their advertised magnitude. All selected Vultr rows have one disk. These nominal values are suitable for coarse catalog comparisons, rather than claims of byte-exact cross-provider equivalence.

Candidate tiers are frozen before examining held-out results. Each tier includes its upper bound and excludes its lower bound; input values must be positive. CPU has 16 tiers, RAM 16, and storage 8. RAM nominal thresholds are 0.5, 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 64, 128, 256, 512, and infinity. CPU thresholds are 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 256, and infinity. Storage thresholds are 32, 64, 128, 256, 512, 1024, 2048, and infinity.

The codec preserves tier values exactly. Precision reduction occurs once when mapping advertised capacities to these tiers. A comparison using exact nominal CPU/RAM/disk triples is reported separately; dictionary coding itself introduces no additional precision loss.

## Construction and independent test

DigitalOcean and Linode supply the dictionary. Identical capacity triples are deduplicated within each provider for ranking. Candidate entries are ranked by provider coverage, distinct-plan count, and tuple order, with a capacity of 64 entries. The resulting table contains 61 entries and leaves three indices unused. The table is frozen before evaluating Vultr and Lightsail.

The dictionary is bound to a proposed outer format version. The measured representation adds one selector bit, uses a six-bit dictionary index on a hit, and falls back to the previous hybrid on a miss. Published entries would remain immutable; unused entries must be rejected until a version explicitly defines them.

| Dataset | Rows | Hits | Hit rate | Hybrid mean label | Dictionary mean label |
| --- | ---: | ---: | ---: | ---: | ---: |
| DigitalOcean (construction) | 31 | 31 | 100.0% | 60.000 | 58.000 |
| Linode (construction) | 48 | 48 | 100.0% | 60.000 | 58.000 |
| Vultr (held out) | 37 | 27 | 73.0% | 60.000 | 58.541 |
| Lightsail (held out) | 11 | 5 | 45.5% | 60.000 | 59.091 |
| Held-out total | 48 | 32 | 66.7% | 60.000 | 58.667 |

Length tests repeat each catalog row with 16 deterministic pseudorandom Node IDs and IP pairs. Both codecs receive identical non-resource fields, including synthetic GeoNames ID 1668341 and Provider ID 42. All resource fields and both IPs are selected. Labels include a one-character prefix and one-character version. The exact character savings depend on those choices and Base36 rounding. Encryption and compact/padded framing have not been integrated.

Hits save ten bits and misses add one bit. Held-out data saves an average 6.33 bits, or 1.33 characters for the tested full records. Deduplicating equal capacities within each held-out provider gives 24/38 hits (63.2%). Exact nominal triples match 18/48 held-out rows (37.5%); the higher tier-based coverage depends on the documented coarser resource semantics. Interpreting held-out advertised disk GB as decimal and converting to GiB changes tier-based coverage to 33/48 (68.8%), illustrating unit sensitivity.

## Robustness and limits

Leave-one-provider-out checks rebuild a table from the other three providers:

| Held-out provider | Table entries | Hits / rows |
| --- | ---: | ---: |
| DigitalOcean | 51 | 9 / 31 |
| Linode | 50 | 19 / 48 |
| Vultr | 64 | 28 / 37 |
| Lightsail | 64 | 6 / 11 |

The differing results demonstrate provider-specific distributions. The initial 66.7% result should not be generalized to every provider or future deployment.

All 117,045 resource-state tuples round-tripped with the frozen table, including absent and zero-valued fields. All 512 optional-field masks at maximum numeric values also round-tripped and stayed within 63 characters. For the held-out CPU/RAM/storage-only scenario, no five-field table entry matched: every dictionary entry includes architecture and environment. The fallback then costs one bit. Independent field selection remains an important reason to retain the general codec.

## Recommendation

Real provider data supports retaining dictionary coding as a serious candidate for full resource descriptions. A table of 61 tier combinations provides measurable cross-provider savings and occupies 122 bytes as packed 16-bit values, before runtime indexing and container overhead. A final format decision should consider sparse selection, ARM and container samples, stable resource semantics, and version maintenance. Evaluating a CPU/RAM/storage-only dictionary could improve portability across architectures and node types, but its mode and presence overhead must be measured separately.

For compact output, successful lookup can shorten a name. For padded output, it frees internal capacity while maintaining the fixed output length. These results concern encoding size and reversibility; security, runtime latency, and production memory use remain unmeasured.

## Frozen candidate table

All entries use architecture code 0 and environment code 1 in this experiment. Ranges are nominal capacity units as described above. Numeric physical capacities remain available in the local normalized source snapshot. Indices 61-63 are unassigned.

| Index | CPU interval | RAM interval | Storage interval | Construction providers |
| --- | --- | --- | --- | --- |
| 0 | (0, 1] | (0.5, 1] | (0, 32] | DigitalOcean, Linode |
| 1 | (0, 1] | (1, 2] | (32, 64] | DigitalOcean, Linode |
| 2 | (1, 2] | (3, 4] | (64, 128] | DigitalOcean, Linode |
| 3 | (3, 4] | (6, 8] | (128, 256] | DigitalOcean, Linode |
| 4 | (6, 8] | (12, 16] | (256, 512] | DigitalOcean, Linode |
| 5 | (24, 32] | (128, 256] | > 2048 | DigitalOcean, Linode |
| 6 | (32, 48] | (64, 128] | (512, 1024] | DigitalOcean, Linode |
| 7 | (48, 64] | (128, 256] | > 2048 | Linode |
| 8 | (0, 1] | (0, 0.5] | (0, 32] | DigitalOcean |
| 9 | (1, 2] | (1, 2] | (32, 64] | DigitalOcean |
| 10 | (1, 2] | (3, 4] | (0, 32] | DigitalOcean |
| 11 | (1, 2] | (3, 4] | (32, 64] | Linode |
| 12 | (1, 2] | (6, 8] | (0, 32] | DigitalOcean |
| 13 | (1, 2] | (6, 8] | (64, 128] | Linode |
| 14 | (1, 2] | (12, 16] | (32, 64] | DigitalOcean |
| 15 | (1, 2] | (12, 16] | (256, 512] | DigitalOcean |
| 16 | (1, 2] | (16, 24] | (0, 32] | Linode |
| 17 | (1, 2] | (32, 64] | (32, 64] | Linode |
| 18 | (3, 4] | (6, 8] | (32, 64] | DigitalOcean |
| 19 | (3, 4] | (6, 8] | (64, 128] | Linode |
| 20 | (3, 4] | (12, 16] | (32, 64] | DigitalOcean |
| 21 | (3, 4] | (12, 16] | (128, 256] | Linode |
| 22 | (3, 4] | (24, 32] | (64, 128] | DigitalOcean |
| 23 | (3, 4] | (24, 32] | (512, 1024] | DigitalOcean |
| 24 | (3, 4] | (64, 128] | (64, 128] | Linode |
| 25 | (4, 6] | (12, 16] | (256, 512] | Linode |
| 26 | (6, 8] | (12, 16] | (64, 128] | DigitalOcean |
| 27 | (6, 8] | (12, 16] | (128, 256] | Linode |
| 28 | (6, 8] | (24, 32] | (64, 128] | DigitalOcean |
| 29 | (6, 8] | (24, 32] | (256, 512] | Linode |
| 30 | (6, 8] | (24, 32] | (512, 1024] | Linode |
| 31 | (6, 8] | (32, 64] | (128, 256] | DigitalOcean |
| 32 | (6, 8] | (32, 64] | (1024, 2048] | DigitalOcean |
| 33 | (6, 8] | (128, 256] | (128, 256] | Linode |
| 34 | (12, 16] | (24, 32] | (128, 256] | DigitalOcean |
| 35 | (12, 16] | (24, 32] | (256, 512] | Linode |
| 36 | (12, 16] | (24, 32] | (512, 1024] | Linode |
| 37 | (12, 16] | (32, 64] | (128, 256] | DigitalOcean |
| 38 | (12, 16] | (32, 64] | (512, 1024] | Linode |
| 39 | (12, 16] | (32, 64] | (1024, 2048] | Linode |
| 40 | (12, 16] | (64, 128] | (256, 512] | DigitalOcean |
| 41 | (12, 16] | (64, 128] | > 2048 | DigitalOcean |
| 42 | (12, 16] | (256, 512] | (256, 512] | Linode |
| 43 | (16, 24] | (64, 128] | (512, 1024] | Linode |
| 44 | (16, 24] | (64, 128] | (1024, 2048] | Linode |
| 45 | (16, 24] | (64, 128] | > 2048 | Linode |
| 46 | (16, 24] | (128, 256] | (512, 1024] | DigitalOcean |
| 47 | (16, 24] | (128, 256] | > 2048 | DigitalOcean |
| 48 | (24, 32] | (32, 64] | (256, 512] | DigitalOcean |
| 49 | (24, 32] | (32, 64] | (512, 1024] | Linode |
| 50 | (24, 32] | (32, 64] | (1024, 2048] | Linode |
| 51 | (24, 32] | (64, 128] | (256, 512] | DigitalOcean |
| 52 | (24, 32] | (64, 128] | (1024, 2048] | Linode |
| 53 | (24, 32] | (128, 256] | (512, 1024] | DigitalOcean |
| 54 | (32, 48] | (64, 128] | (1024, 2048] | Linode |
| 55 | (32, 48] | (128, 256] | (256, 512] | DigitalOcean |
| 56 | (48, 64] | (64, 128] | (1024, 2048] | Linode |
| 57 | (48, 64] | (64, 128] | > 2048 | Linode |
| 58 | (96, 128] | (128, 256] | > 2048 | Linode |
| 59 | (96, 128] | (256, 512] | > 2048 | Linode |
| 60 | (128, 256] | (256, 512] | > 2048 | Linode |

Table SHA-256: `3c14508fb22d7b6a8f1bcb80cab82ee855f8c4fd21abf5ef52890c676d29381e`.

## Source references

- [DigitalOcean](https://www.digitalocean.com/pricing/droplets)
- [Linode](https://api.linode.com/v4/linode/types?page_size=500)
- [Vultr](https://api.vultr.com/v2/plans?per_page=500)
- [Lightsail](https://aws.amazon.com/lightsail/pricing/)
