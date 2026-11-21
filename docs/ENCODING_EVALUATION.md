# Hostname Format Evaluation

Date: 2026-11-19. Status: evaluation and recommendations; protocol selection remains open.

## Objective

Produce consistent node hostnames that carry selected node information. Preserve an 80-bit identity, independent field selection, reversible identifiers and addresses, resource tiers, and a single label of at most 63 ASCII characters. Evaluate both compact encoding and operational naming behavior.

## Naming approaches

| Approach | Benefits | Tradeoffs |
| --- | --- | --- |
| Versioned self-contained descriptor | Decodes selected metadata without per-node storage; configurable confidentiality | Attribute changes can change the hostname; strict capacity budget |
| Readable location/provider plus random identity | Easy visual recognition and consistent naming | Exposes metadata; readable fields consume more characters; registry semantics require maintenance |
| Stable random hostname with metadata in DNS TXT or a service | Short stable endpoint; metadata can evolve independently | Requires metadata publication, access policy, and lookup infrastructure |

The self-contained descriptor best preserves the established requirements. A stable alias can be added later if permanent service endpoints become necessary. Metadata selection remains explicit; a small default selection can keep ordinary names short.

## Candidate outer envelope

Use `n<version><payload>.<configured-domain>`, with one lowercase Base36 character for the visible envelope version. This provides 36 symbols; an extension symbol can introduce a longer future header with a correspondingly smaller payload budget. Each version specifies its packing, encryption behavior, and canonical representation. Product release versions remain independent.

The experiment uses `n1` as a two-character envelope and leaves 61 payload characters. Moving the original three version bits out of the packed content leaves a 314-bit maximum. A one-bit leading sentinel preserves bit length and leading zeros, producing at most 315 bits. Every 315-bit value fits in 61 Base36 characters because `2^315 < 36^61`.

This is a clearer decoding entry point than an encrypted format discriminator. Visible version information also reveals the format family. Encryption would preserve the exact payload string length, including any leading zero characters in ciphertext.

An independent key selector has not been budgeted. Arbitrarily adding another header character may exceed the full-field budget. Key rotation requires an explicit external key selection or a budgeted selector/envelope revision. Unauthenticated decryption must never choose a key merely because its result parses successfully.

## Offline experiment

A Python evaluation script in local ignored workspace state exercised all 512 optional-field combinations with zero, small, maximum, and pseudorandom values, plus 1,000 additional pseudorandom inputs. All 3,048 inputs passed six packing round trips each, for 18,288 round-trip assertions. Zero values and absent fields remain distinct.

The experiment compares fixed packing with a four-mode hybrid:

- Fixed-width identifiers and a sparse resource bitmap.
- Fixed-width identifiers and a 17-bit mixed-radix resource representation.
- Unsigned varint identifiers and a sparse resource bitmap.
- Unsigned varint identifiers and the 17-bit resource representation.

The hybrid includes a two-bit mode selector and chooses the smallest bit representation, breaking ties by mode number. Every mode retains the original integer ranges. Raw IPv4/IPv6 and the random identity retain their complete widths.

| Input | Original packing with sentinel | Visible version, fixed packing | Visible version, hybrid |
| --- | ---: | ---: | ---: |
| Identity only | 19 | 19 | 19 |
| Identity, example geography and provider | 28 | 28 | 26 |
| Identity and all resource fields | 23 | 23 | 23 |
| Identity and example IPv6 | 43 | 44 | 44 |
| Identity and both example IPs | 50 | 50 | 50 |
| All example fields | 63 | 63 | 60 |
| All fields at maximum values | 63 | 63 | 63 |

Lengths include the prefix and version where applicable, and exclude the domain suffix. Example values include geographic ID 1668341, provider 42, and documentation-range IP addresses. These examples are synthetic.

Across the synthetic test distribution, the hybrid averaged 43.86 characters versus 44.63 for fixed packing. It was shorter for 805 inputs, equal for 1,652, and longer for 591. Both had a maximum of 63 characters. These averages describe the test corpus, not expected production traffic.

CBOR and Protobuf were evaluated with byte-size models, not installed library encoders. The CBOR model uses the smaller of a positional array with nulls and an integer-key map. The Protobuf model uses compact field numbers, optional integer fields, and binary addresses. For the full example they require 46 and 51 bytes respectively, before encryption or additional framing; conservative Base36 label bounds with the two-character envelope are 74 and 81 characters. These results apply to the modeled schemas, rather than every possible use of those formats.

Standard zlib compression was also exercised on packed bytes. It expanded small records and benefited highly repetitive values; the full example required a 64-character conservative label bound. It does not establish a universal capacity improvement.

## Recommendation and limits

Use a versioned self-contained hostname as the main design candidate. Adopt the visible outer version as a candidate envelope. Retain fixed packing as the baseline; evaluate whether a typical two-to-three-character saving warrants the hybrid's extra decoding paths before freezing the first format.

Field-specific normalization and representation are appropriate optimization targets. Persistent dictionaries for individual nodes would introduce lookup state. Hashing metadata would require an external source to recover the original values.

This experiment validates reversible packing and size only. It does not validate FF1 encryption, cryptographic security, runtime performance, extension handling, or production canonical-input rejection. Encryption selection remains open. The inspected Capital One FPE repository is archived and its README explicitly states that it is unmaintained; it should not be adopted solely because it implements FF1. Dependency evaluation must cover maintenance, license notices, applicable algorithm requirements, and official test vectors.

Before implementation, settle the resource code tables, provider dictionary semantics, encryption implementation, key lifecycle, and the default selected metadata. Measure runtime and memory with the eventual Go implementation.

## References

- [DNS label representation](https://www.rfc-editor.org/rfc/rfc1035)
- [CBOR](https://www.rfc-editor.org/rfc/rfc8949)
- [Protobuf encoding](https://protobuf.dev/programming-guides/encoding/)
- [Capital One FPE repository](https://github.com/capitalone/fpe)

Source details and dependency status should be rechecked when selecting implementation dependencies.

## Resource Dictionary Experiment (2026-11-20)

The subsequent [provider-catalog evaluation](RESOURCE_DICTIONARY_EVALUATION.md) builds a concrete table from official plan data and tests independent providers. The synthetic experiment below remains a reference for controlled hit-rate scenarios.

A second isolated experiment adds a 64-entry resource dictionary with a fallback to the four-mode hybrid. The table is frozen before corpus generation. It contains two architecture codes, two environment codes, and sixteen correlated CPU/RAM/storage tuples. These are illustrative numeric codes using the original field widths; physical tier meanings and the proposed node-type replacement remain undecided. The table has no claim to represent real deployment frequencies.

The encoded payload includes one selector bit. A dictionary hit stores a six-bit index and encodes the remaining fields with resources absent. A miss uses the existing hybrid with the extra selector. The dictionary would be bound to the outer format version; no separate dictionary identifier is included in these measurements. Published mappings must remain immutable.

For full resource sets, a hit saves ten bits relative to the previous hybrid, including selector overhead. A miss adds one bit. At the bit level the full-resource break-even hit rate is greater than `1/11`, or approximately 9.1%. Base36 rounding affects actual character savings.

| Synthetic full-resource hit rate | Previous hybrid mean characters | Dictionary mean characters |
| --- | ---: | ---: |
| 0% | 62.531 | 62.740 |
| 10% | 62.543 | 62.549 |
| 50% | 62.558 | 61.742 |
| 90% | 62.569 | 60.929 |
| 100% | 62.575 | 60.725 |

Each row uses 1,000 independently generated records with all other optional fields present. Rates are deliberately constructed scenarios. Independent deterministic samples introduce small baseline differences between rows. The original mask/random corpus yielded 16 hits among 3,048 records; average length increased from 43.855 to 43.975 characters. A corpus selecting one resource field at a time had no hits. Compact label lengths include a one-character prefix and one-character version and exclude the domain suffix.

Validation passed for all 117,045 resource-state combinations, including absence and numeric zero, and all 512 optional-field masks at maximum values. Dictionary hits and fallback records round-tripped exactly. Maximum boundary length remained 63 characters. The table contains 128 bytes of packed resource values; this excludes runtime containers, lookup indices, and code. Runtime and memory were not benchmarked.

The experiment covers reversible packing and length. Encryption, compact/padded discrimination, production input rejection, and final field semantics remain outside the tested format. In padded output, compression would free internal capacity while the visible label remains 63 characters.

Recommendation: retain dictionary coding as an optional format candidate until representative node configurations and final resource tiers establish a worthwhile hit rate. Preserve independent field selection in the interface. The implementer can select dictionary or fallback encoding automatically within a chosen format. Existing fixed/hybrid formats remain available without paying a dictionary selector cost. No production codec or protocol selection was changed by this experiment.
