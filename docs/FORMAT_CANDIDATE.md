# Descriptor Format Candidate and Encryption Evaluation

Date: 2026-11-20. Historical evaluation record. The frozen production wire specification is [Descriptor Format 1](FORMAT_1.md); this document preserves the design work that led to formats 0 and 1.

## Proposed field semantics

Node identity uses 80 CSPRNG-generated bits. Geographic identity uses a 32-bit GeoNames ID. IPv4 and IPv6 use their complete 32-bit and 128-bit values. Users retain Node ID when updating attributes. Resource values describe assigned capacity, with a stable interval displayed before encoding.

| Field | Proposed code meanings |
| --- | --- |
| Architecture, 2 bits | 0: x86-64; 1: ARM64; 2: RISC-V 64; 3: other |
| Node type, 3 bits | 0: physical machine; 1: virtual machine; 2: system container; 3: application container; 4-6: reserved; 7: other |
| CPU capacity, 4 bits | Inclusive upper bounds: 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 256, infinity |
| RAM capacity, 4 bits | Inclusive GiB upper bounds: 0.5, 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 64, 128, 256, 512, infinity |
| Storage capacity, 3 bits | Inclusive GiB upper bounds: 32, 64, 128, 256, 512, 1024, 2048, infinity |

The tier code is the zero-based index in its ordered list. The first interval starts above zero; later intervals exclude the previous bound and include their own bound. CPU capacity represents assigned logical CPUs or a positive fractional CPU quota. RAM represents assigned memory. Storage represents the capacity selected by the user under a documented local-storage accounting rule; live free space is outside that value. Decimal input units must be converted explicitly before tier selection. An absent field remains distinct from code zero. Unknown capacities can remain unselected.

These are proposals derived from the catalog experiment. The original prototype used a generic three-bit environment value; the proposed node-type meanings are a semantic refinement that requires dedicated production validation. The codec experiments tested the numeric domain, including values now proposed as reserved.

## Provider registry proposal

Retain a 16-bit identifier backed by a versioned public registry. Proposed initial entries are 0: self-managed, 1: DigitalOcean, 2: Akamai Cloud/Linode, 3: Vultr, 4: AWS, and 65535: other. Entries 1-4 reuse the evaluation's provisional mapping; adoption would explicitly freeze their meanings. Provider identifies the hosting source; product-family labels remain reference metadata. Registry additions preserve existing identifiers permanently.

An operator's arbitrary local aliases must not redefine public IDs. Unknown numeric IDs can remain visible as unresolved identifiers; a lookup failure preserves the encoded value. The other category carries a category code, with free-text operator notes remaining outside the descriptor. Provider-specific menus and recommendations can evolve independently of immutable decoding dictionaries.

## Codec direction

The raw bitmap codec remains the reference. The generic projected CPU/RAM/storage dictionary remains the preferred compact candidate from the [codec comparison](CODEC_COMPARISON.md). Architecture and node type remain independently selected. Provider-specific UI recommendations can guide input without requiring the corresponding provider-specific wire table.

Field-pattern coding remains provisional because measured user-selection frequencies are unavailable. The encryption integration below uses the tested `patterns_generic` candidate; it demonstrates complete cryptographic round trips without freezing that candidate as the production format.

## Encryption evaluation

Two upstream implementations were downloaded at fixed commits into isolated local evaluation storage:

| Implementation | Commit | License | Role |
| --- | --- | --- | --- |
| joshuajorel/scrambler, ff1 package | 7a8109c8d38f8daac7914f2d3dba63ee52ba2c16 | Apache-2.0 | Primary evaluation candidate |
| ubiqsecurity/ubiq-go, structured package | 2dd650b7dcfb872422b47f4d95ac8d7b29098670 | MIT | Independent output comparison |

Scrambler's FF1 package uses standard-library dependencies and integer arithmetic for domain limits and FF1 calculations. The selected package exposes per-call tweaks and configurable input limits. Ubiq's FF1 context contains mutable arithmetic state, and its minimum-length calculation uses floating-point logarithms; it was used sequentially as a comparison implementation. Its structured package requires golang.org/x/exp/slices. These checks are targeted source inspection, not a security audit or endorsement based on repository recency.

The experiment used radix 36, the alphabet `0123456789abcdefghijklmnopqrstuvwxyz`, AES-256 test keys, and an explicit fixed evaluation tweak. Keys were public deterministic fixtures, with no deployment secrets. Each record included Node ID and all selected metadata in the encrypted payload. Input limits were 17-61 characters and a maximum 64-byte tweak. The application limit of 17 is derived from the shortest tested candidate payload; a changed production layout must recalculate it.

Validation results:

- Selected upstream NIST example, ACVP vector, minimum-length, invalid-symbol, and concurrent-use tests passed on Windows.
- All 3,072 PND fixtures matched between the two FF1 implementations, including compact and padded encodings, zero/maximum/random values, and every optional-field mask.
- The complete sequence of packing, encryption, decryption, and unpacking recovered every original record exactly.
- Ciphertext length remained equal to plaintext length. All labels stayed within 63 characters; padded labels were exactly 63.
- Seventy-seven ciphertexts began with zero and round-tripped correctly. Leading ciphertext zeros are significant and must be retained until decryption.
- Repeated encryption with the same parameters was deterministic. A different test key changed all 3,072 tested outputs; the original key still recovered original outputs.
- Four invalid length/alphabet inputs were rejected by the primary cipher API. Wrong-key or altered-ciphertext rejection is not an authenticity guarantee.
- The primary FF1 test binary cross-compiled for Linux amd64. Linux execution and race-instrumented testing were not performed in this step.

A local Windows microbenchmark measured approximately 1.27 microseconds per 17-character encryption and 3.05 microseconds per 61-character encryption, with 45 and 68 allocations respectively. It used 10,000 iterations and excluded cipher construction, HTTP, SSO, reference lookup, and process memory. These are local observations, not production performance budgets.

The first implementation is a suitable candidate for further review. Its Apache-2.0 license remains applicable to that dependency and requires preservation of applicable license and notice material. PND's own code is planned under MIT. Final dependency adoption includes license/NOTICE review, pinned versions, and a documented upgrade policy. Passing vectors and differential tests establishes observed agreement, not independent cryptographic assurance.

## Adopted MVP envelope and key lifecycle

The label uses one configurable lowercase alphabetic prefix, one format-version character, and up to 61 ciphertext characters. Version `0` identifies the MVP codec and immutable dictionary. The server loads one 32-byte key from a file at startup. Changing that file and restarting replaces the active key. Names contain no key-generation identifier or key-profile mapping.

Existing descriptors require the original key for correct recovery. Operators retain exported node data to regenerate names after key replacement and update DNS and downstream URLs. The server performs structural validation after decryption; a wrong key can still produce plausible data because descriptor authentication is outside this format.

The fixed format-0 tweak is `PND:format-0:ff1:aes256`. Prefix and domain changes preserve payload decodability. The Node ID remains inside the encrypted payload.

## Compact and padded decoding

Compact mode encrypts the shortest canonical Base36 representation. Padded mode left-pads that representation to 61 characters before encryption. After selecting the format and decrypting with the active key, the decoder checks canonical padding, removes it numerically, and recovers the original bit length from the sentinel. No manual input-mode selection or padding-mode bit is required. Compact values that already use 61 characters coincide with their padded representation.

Both modes reveal repeated inputs under the same key and tweak. Padding hides the original length; it does not create new entropy. Descriptors carry descriptive data and are not authorization credentials. Public decoder access exposes the supported contents to its callers.

## Adopted format and release verification

The release candidate adopts pattern-assisted framing with the projected generic resource dictionary. Field semantics, tier intervals and Provider IDs are recorded in TECHNICAL.md and the codec catalog. Independent packing vectors, all optional-field masks and key-change tests cover the implemented format.

Preserve the single-key startup contract and format-0 dictionary meanings. Complete the release verification in RELEASING.md before publishing a stable tag; subsequent interpretation changes require a new descriptor version.

## Sources

- [Scrambler source at evaluated commit](https://github.com/joshuajorel/scrambler/tree/7a8109c8d38f8daac7914f2d3dba63ee52ba2c16)
- [Ubiq source at evaluated commit](https://github.com/ubiqsecurity/ubiq-go/tree/2dd650b7dcfb872422b47f4d95ac8d7b29098670)
- [NIST FF1 examples](https://csrc.nist.gov/CSRC/media/Projects/Cryptographic-Standards-and-Guidelines/documents/examples/FF1samples.pdf)
- [NIST SP 800-38G Revision 1 second public draft](https://csrc.nist.gov/pubs/sp/800/38/g/r1/2pd)

The referenced revision is a draft; its requirements are distinguished from the earlier published standard. Current standards status and dependency versions must be verified again at production selection.
