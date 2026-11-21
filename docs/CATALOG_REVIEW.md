# Provider Catalog Review

Prepublication review date: 2026-11-21. Candidate: `26.1.0-rc.2`.

The snapshot includes 38 hosting providers, Self-managed and Other, with 587 recommendations across 62 product series. Provider search supports names, aliases and IDs. Plan search, series selection, matching-field filtering and pagination expose the reviewed specifications and source links. Applying a plan preserves identity, location and network values, and preserves separately configured storage.

Validation passed for generated data, decimal/binary unit boundaries, Oracle OCPU conversion, provider coverage and the frozen format-0 dictionary. Go tests passed on Windows; focused catalog, codec and server tests passed on Linux. Local and remote browser checks passed for the new catalog and the existing application workflows. The shared browser demo also passed locally; the public Pages site is unchanged during review.

A synthetic descriptor was encoded before the deployment update. The updated service produced the identical hostname and decoded the existing name into the same fields. The deployment retained its key, hostname defaults, GeoNames account configuration and GeoIP database. One post-browser-test sample reported 18.38 MiB memory and 0.00% CPU against the existing 128 MiB limit. This is an observation from the review host, not a peak-memory guarantee.

The candidate was verified on the LAN review service before publication. Catalog source coverage and contribution steps are documented in [Catalog](CATALOG.md).

The follow-up revision adds OneProvider, DediRock and a broader VPS/dedicated-hosting directory. Searchable provider controls now match native select backgrounds, borders, font sizes and heights in both themes. The recommendations disclosure has a full-width bordered trigger and rotating chevron. Browser checks verify visual style parity and directory-only provider encoding.

All single-select controls now share one searchable popup and native listbox implementation, including resource fields, country/city helpers, product series, language, theme and length mode. Other/self-managed/system choices have separate labeled groups where applicable. Tests cover keyboard selection, Escape, disabled states, retained values, consistent light/dark control surfaces and mobile popup placement.

Product-series filters identify providers across the catalog and separate reference configurations from product families. Disclosure panels use immediate layout with a short menu fade, and boolean controls share consistent inactive and active styles. Desktop/mobile geometry checks and light/dark toggle checks passed on the review deployment.
