# PND Descriptor Format 1

Status: frozen wire specification for the 26.1.0 release line. New encodings use format 1. Format 0 remains supported for decoding with the same deployment key. Product versions and descriptor versions evolve independently.

## Envelope and input

A generated hostname is `<prefix>1<ciphertext>.<domain>`. The prefix is one ASCII letter `a` through `z`. The complete descriptor label is 19 to 63 ASCII characters. Ciphertext uses the ordered alphabet `0123456789abcdefghijklmnopqrstuvwxyz`, with 17 to 61 characters. Its leading zero characters are significant.

The configured domain contains lowercase ASCII DNS labels, each 1 to 63 characters, using letters, digits and internal hyphens. The suffix is at most 189 characters, giving a maximum 253-character hostname with a 63-character descriptor label. Prefix and domain are naming settings and do not enter encryption.

The decoder accepts a label, hostname, or HTTP(S) URL. It trims surrounding whitespace, caps the resulting input at 2048 bytes, extracts the URL hostname when applicable, then takes the first DNS label and folds ASCII case. URL user information is rejected. Versions other than `0` and `1` are rejected. Selecting the version also selects the cryptographic tweak; migration recovers the descriptor with its original version and encodes a new name with the target version.

## Fields and units

Node ID is mandatory: ten bytes, represented in JSON as 20 hexadecimal characters. New identities use a cryptographically secure random generator. Updating a node retains its ID. All other fields are independently optional. JSON omission and a selected numeric value of zero have distinct meanings.

| Mask bit (LSB = 0) | JSON field | Encoding width | Meaning |
| --- | --- | ---: | --- |
| 0 | `geo` | 32 bits or variable integer | GeoNames ID, 1 through 4294967295 |
| 1 | `provider` | 16 bits or variable integer | Provider registry ID, 0 through 65535 |
| 2 | `architecture` | 2 bits | 0 x86-64; 1 ARM64; 2 RISC-V 64; 3 other |
| 3 | `nodeType` | 3 bits | 0 physical; 1 VM; 2 system container; 3 application container; 7 other |
| 4 | `cpu` | 4 bits or dictionary index | Assigned logical CPU count or positive CPU quota interval |
| 5 | `ram` | 4 bits or dictionary index | Assigned memory interval in GiB |
| 6 | `storage` | 3 bits or dictionary index | Selected local-storage capacity interval in GiB |
| 7 | `ipv4` | 32 bits | IPv4 address in network byte order |
| 8 | `ipv6` | 128 bits | IPv6 address in network byte order |

Node-type values 4, 5 and 6 are reserved and rejected. IPv6 zone identifiers and IPv4-mapped IPv6 addresses are rejected; IPv4 has its own field. Addresses are normalized using their numeric values. The format stores configured capacity; operators choose which local volumes contribute to storage. Free space, prices and traffic limits stay outside these fields.

Resource codes are zero-based indices into these immutable inclusive upper bounds:

| Field | Upper bounds, in code order |
| --- | --- |
| CPU | 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 48, 64, 96, 128, 256, infinity |
| RAM (GiB) | 0.5, 1, 2, 3, 4, 6, 8, 12, 16, 24, 32, 64, 128, 256, 512, infinity |
| Storage (GiB) | 32, 64, 128, 256, 512, 1024, 2048, infinity |

The first interval excludes zero. Later intervals exclude the previous bound and include their own bound. Decimal GB uses 10^9 bytes; GiB uses 2^30 bytes. Convert units before assigning a tier. Decoding recovers the interval code.

Provider IDs are defined in the [public registry](CATALOG.md#provider-ids). Existing IDs retain their meanings permanently. Additions allocate unused values. Unrecognized numeric IDs remain decodable. Provider names, aliases, recommendations, geography labels and available UI options can change independently of the wire format.

## Bitstream

Append bits from left to right, most-significant bit first within each fixed-width integer and byte. Concatenate the following sections without byte alignment:

1. One sentinel bit with value `1`.
2. The presence header described below.
3. The ten Node ID bytes in hexadecimal display order.
4. Geographic/provider integers, when selected.
5. Architecture followed by node type, when selected.
6. Resource section, when selected.
7. IPv4 bytes followed by IPv6 bytes, when selected.

### Presence header

Compute `mask = sum(1 << bit)` for the selected fields. The immutable pattern table is:

| Index | Mask | Index | Mask |
| ---: | ---: | ---: | ---: |
| 0 | 0 | 8 | 387 |
| 1 | 3 | 9 | 511 |
| 2 | 124 | 10 | 507 |
| 3 | 112 | 11 | 255 |
| 4 | 114 | 12 | 383 |
| 5 | 126 | 13 | 1 |
| 6 | 127 | 14 | 2 |
| 7 | 384 | 15 | 128 |

For a table match, append `1` and the four-bit index. Otherwise append `0` and the nine-bit mask. A raw mask that matches a pattern is noncanonical and rejected.

### Geographic/provider integers

When either integer is present, append one shared selector. Compute the sum of the selected fixed widths (32 for GeoNames ID, 16 for Provider ID) and the sum of their shortest unsigned LEB128 widths. Choose selector `1` only when LEB128 uses strictly fewer bits; ties use selector `0` and fixed widths. Then append the selected geographic ID and provider ID in that order using the chosen representation.

Unsigned LEB128 emits low seven-bit groups first. The high bit of every byte is `1` when another group follows, and `0` in the last byte. Each byte is appended MSB-first to the unaligned bitstream. Values must fit their field's numeric range. Redundant groups are noncanonical.

### Resource section

The normative dictionary is [dictionary.json](../internal/catalog/dictionary.json), containing 73 triples in CPU/RAM/storage code order. Formats 0 and 1 share this immutable dictionary. Its SHA-256 after compact JSON serialization of the integer arrays is:

```text
744533032e5282766388956d37298cd41f2be0c6ba9e3a999cc897d03db0bed3
```

Project the triples onto the selected resource columns in CPU/RAM/storage order. Deduplicate identical projected tuples, then sort them lexicographically by numeric element values. For N entries, the dictionary index width is `ceil(log2(N))`, with zero bits for N = 1.

When the selected tuple occurs in this projection and its index width is strictly smaller than the sum of the selected direct widths, append selector `1` and its zero-based index. Otherwise append selector `0` and the selected direct codes in CPU/RAM/storage order. A tie chooses direct codes. No selector is emitted when no resource fields are selected. Unused indices and noncanonical encoding choices are rejected.

## Base36, padding and encryption

Interpret the complete sentinel-prefixed bitstream as a positive integer and serialize its shortest lowercase Base36 representation. Compact mode uses that string. Padded mode prepends zero characters until it has exactly 61 characters. Both use the same decoder. If the compact representation already has length 61, the two outputs coincide.

Encrypt the entire string using FF1 with AES-256, radix 36 and the alphabet above. The key is exactly 32 bytes. The format-1 tweak is the ASCII byte string `PND:format-1:ff1:aes256`, with no terminator. Format 0 uses `PND:format-0:ff1:aes256`. Ciphertext has the same length as plaintext. Both layouts fit within the 315-bit capacity budget; the direct-field upper bound for the frozen field patterns is 314 bits including the sentinel.

On decode, select the tweak from the visible format character and decrypt before removing padding. The plaintext must be either the shortest nonzero Base36 representation or a 61-character representation with leading zero padding. Convert it to binary, remove its highest set bit (the sentinel), and parse the remaining fields. Reject truncation, trailing data, reserved values and invalid addresses. Repack the descriptor and require the shortest Base36 result to match, enforcing every canonical choice above.

The service loads one key at startup. Replacing the key requires retaining old node exports to generate replacement names. Identical inputs, key, format and length mode produce identical names. Padding conceals the original length. The format provides confidentiality; cryptographic message authentication is outside its scope. Modified ciphertext or an incorrect key may decode to plausible fields. Applications use their own authorization and trust policy for decoded values.

## Compatibility and conformance

Format 1 freezes the existing layout with a distinct version and tweak. New output uses format 1; format-0 input remains supported with its original tweak and dictionary. Reusing a decoded format-0 descriptor produces a format-1 name and preserves its Node ID. Keep existing DNS names until replacements are deployed where needed. Released rc.1 through rc.3 applications decode format 0; deploying format-1 names requires the new decoder on their consumers.

Fixed public-key vectors are distributed in [format0.json](../internal/codec/testdata/format0.json) and [format1.json](../internal/codec/testdata/format1.json). They cover compact and padded output with two synthetic AES keys. Independent packing vectors are in [packing.json](../internal/codec/testdata/packing.json). Public fixture keys are for testing.

Conformance requires exact ciphertext agreement, both-version decoding, all 512 optional-field masks, numeric boundary validation, canonical encoding, and Windows/Linux/browser-WASM parity. Fuzz tests exercise malformed input and both-version round trips. Changing field meanings, pattern order, dictionary values, canonical selection or encryption parameters requires a new format version. UI labels and reference catalogs retain their independent update lifecycle.

FF1 is specified in [NIST SP 800-38G](https://csrc.nist.gov/pubs/sp/800/38/g/final). The [revision 1 second public draft](https://csrc.nist.gov/pubs/sp/800/38/g/r1/2pd) addresses domain size and implementation arithmetic; its publication page remained a draft at the candidate review. PND's implementation uses integer arithmetic and a minimum 17-character radix-36 domain. This specification records PND interoperability rules; independent cryptographic certification is outside the release validation.
