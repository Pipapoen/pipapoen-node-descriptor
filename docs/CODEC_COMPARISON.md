# PND Codec Candidate Comparison

Date: 2026-11-20. Status: offline evaluation; production protocol and encryption integration remain pending.

## Objective and evidence

Compare selection-pattern coding, provider-specific dictionaries, and conditional CPU/RAM/storage dictionaries while retaining independent optional fields, an 80-bit identity, and a 63-character label. The experiment reuses the 127 official catalog rows documented in [the resource dictionary evaluation](RESOURCE_DICTIONARY_EVALUATION.md). Source capacity normalization and provisional tier boundaries are unchanged. Catalog prevalence and user field-selection frequency remain unknown.

All candidates use the original resource widths, including the three-bit environment field. Provider identifiers 1-4 are experimental mappings, not a public registry. Architecture and environment are encoded independently of resource dictionaries, enabling the same CPU/RAM/storage entries to apply when those fields are omitted or differ. Compression preserves the supplied tiers exactly.

## Candidates

| Candidate | Selection header | Resource values |
| --- | --- | --- |
| Raw baseline | Nine-bit optional-field bitmap | Individual fields |
| Pattern baseline | One flag plus four-bit pattern index, or one flag plus nine-bit bitmap | Individual fields |
| Generic dictionary | Pattern header | One selector bit for generic dictionary or individual fields |
| Provider dictionary | Pattern header | Two-bit selector for individual, generic, provider, or reserved mode |
| Hierarchical dictionary | Pattern header | Two-bit selector; fourth mode uses CPU/RAM then conditional storage |

Each candidate includes a leading-one sentinel, an 80-bit Node ID, a one-bit fixed/varint choice when geographic/provider identifiers exist, selected architecture/environment values, resource values, and complete selected IP addresses. The source widths and all control bits are included in measurements. A one-character prefix and one-character visible format version are included in label lengths.

The pattern table contains sixteen UI-inspired hypotheses, frozen before measurement. It includes all-fields and all-fields-except-architecture patterns to keep expensive near-full selections compact. These patterns are proposed defaults without measured user-frequency evidence. Other combinations use the complete nine-bit bitmap. A pattern reveals exactly which fields are present; dictionaries never add unselected fields.

Resource dictionaries are projections of version-bound CPU/RAM/storage tuples onto the selected resource columns. Each projection is deduplicated and sorted. Its index width is `ceil(log2(entry_count))`; a single known option needs zero index bits. Decoders reject unavailable tables and unused index values. Published dictionaries must remain immutable within their format version.

Hierarchical mode first selects or directly stores the chosen CPU/RAM fields. It then uses those recovered values to select a storage table. Each stage includes a dictionary/raw flag. Provider context is used when an appropriate table exists; otherwise the global table supplies context. The overall encoder compares complete alternatives, selects the fewest bits, and resolves ties by a fixed mode order.

## Results

Mean compact label characters, with the domain suffix excluded:

| Scenario | Raw baseline | Pattern baseline | Generic dictionary | Provider dictionary | Hierarchical dictionary |
| --- | ---: | ---: | ---: | ---: | ---: |
| Known catalog, full selection | 60.000 | 60.000 | 59.000 | 59.000 | 59.000 |
| Held-out providers, full selection | 60.000 | 60.000 | 59.333 | 59.333 | 59.333 |
| Held-out providers, selected sparse scenarios | 30.667 | 29.778 | 29.407 | 29.630 | 29.630 |
| Catalog with every one of 512 selection masks | 40.092 | 40.189 | 40.172 | 40.260 | 40.260 |
| Known catalog, Provider omitted | 59.000 | 59.000 | 58.000 | 59.000 | 59.000 |
| Unseen resource tuples within known providers | 60.000 | 60.000 | 59.636 | 59.636 | 59.636 |

Known-catalog scenarios are capacity demonstrations using the construction data. Held-out-provider scenarios build tables from DigitalOcean and Linode and evaluate Vultr and Lightsail. Within-provider holdout partitions whole resource-tier tuples using a deterministic hash, so identical tuples stay on the same side of the split. That split held out 22 plan rows. Sparse scenarios use nine explicit masks; the exhaustive mask scenario weights all masks equally. Neither approximates measured user behavior.

Every catalog row/mask uses four deterministic random Node ID/IP combinations. Geographic ID 1668341 and small experimental Provider IDs are fixed when selected. Different identifier widths and values can change character rounding. The exhaustive mask scenario contains 260,096 records per candidate.

For known full records, provider dictionaries save an average 0.709 bits over the generic-dictionary candidate, with no character reduction in these samples. Hierarchical mode was selected for 12 of 88 within-provider holdout records, reducing average length by approximately 0.136 bits compared with provider mode alone. Its additional complexity produced no character reduction in that scenario.

The raw baseline performs best when all 512 field-selection masks receive equal weight. Pattern coding helps the explicitly selected sparse scenarios. This difference makes actual selection frequency important before adopting a selection-pattern dictionary.

Compared with the previous four-mode hybrid, the generic candidate improves the known full-record example from 60 to 59 characters and the selected sparse scenario from 30.222 to 29.407. The previous five-field whole-resource dictionary reported different results because it also compressed architecture/environment and used different presence and mode framing. Those results should not be treated as a direct isolated test of provider context.

## Compact and padded framing

The tested pre-encryption framing converts the sentinel-prefixed bitstring to Base36. Compact mode uses its shortest representation. Padded mode left-pads that representation with zero characters to 61 payload characters. A one-character configurable lowercase alphabetic prefix and one version character produce a 63-character padded label.

After decryption, numeric parsing discards the left padding, and the sentinel restores the original bitstring length. The decoder requires either the shortest representation or exactly 61 correctly padded characters. A separate padding-mode bit is unnecessary for this candidate. A compact payload already occupying 61 characters is identical to its padded representation; both recover the same data.

Encryption remains an explicit boundary: a future length-preserving cipher must process the entire payload, retain ciphertext leading zeros, and decrypt before padding removal. No cipher or toy substitute was used in this experiment. Correct pre-encryption framing does not establish encryption safety. Key selection, key rotation, and cipher parameters are still open; an extra key identifier has no reserved capacity in these measurements. Prefix changes are accepted independently of decoding data, while supported format versions determine parsing and dictionary rules.

## Verification

- All catalog scenario records round-tripped for all five candidates.
- All 117,045 resource states round-tripped for the hierarchical candidate with a known provider, and separately for the raw and generic candidates with an unknown provider.
- All 512 optional-field masks were tested with zero, maximum, and random values: 1,536 boundary inputs per candidate.
- Compact/padded and alternate-prefix/case handling passed 15,360 envelope round trips across the five candidates. Twenty-nine hierarchical boundary records had identical compact/padded payloads, as expected at full capacity.
- Seven malformed structure/envelope cases and 100 unassigned dictionary indices were rejected. This is targeted validation rather than a complete malformed-input fuzz campaign or authenticity check.
- The hierarchical candidate reached at most 315 packed bits and 63 label characters. All candidates fit the tested boundary cases. With these widths, the two most expensive selection masks have pattern entries; all other masks omit enough field bits for the raw escape header to fit.
- Unknown Provider IDs recovered correctly through generic/raw fallback. Single-option tables, absent fields, and numeric zero remained distinct where applicable.

The measured table set contains 73 global triples and provider tables of 31, 37, 20, and 11 triples. Its serialized test-table SHA-256 is `568bf82f30a7dee3b3752f852f5d1cb13b4d4ebda137e7a000c7b42c664cc6e6`. The Python experiment is an evaluation harness; its runtime and object memory do not establish production Go performance.

## Recommendation

Use the raw bitmap codec as the reference and retain the generic projected resource dictionary as the preferred compact candidate for MVP validation. Pattern coding should remain provisional until the intended default selections or actual trial usage justify its header cost. The measured provider and hierarchical variants do not yet justify their extra parsing and table-maintenance paths for the first format.

Keep provider-aware dynamic recommendations in the interface independently of the wire codec. They can suggest matching CPU/RAM/storage selections while preserving the user's values, stable tier definitions, custom-entry access, and independent field selection. The wire encoder can use a common table even when the UI recommendation came from a provider-specific catalog.

Before freezing the first production format, finalize resource and provider semantics, choose and validate the encryption implementation, resolve key identification, and implement the chosen codec in Go with shared test vectors. This evaluation changes documentation and isolated experiment artifacts only.
