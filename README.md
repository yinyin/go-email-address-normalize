# Purpose

Provide some level of malformed Email address filtering to protect MTAs.

Validation rules will based on the RFCs but with some additional limitations.

# Result Value

There are 3 values in result:

- `checkedEmailAddress`: Email address with minimal *fixes*.
  In current implementation *local part quoting* is the only fix.
- `normalizedEmailAddress`: Checked email address with normalizations. The normalization including: lower casing, consolidate dots and spaces, remove sub-addressing.
- `err`: Validating errors.

# Validation Rules

## Local Part

* Dot constraints are ignored: dot can appear at anywhere.
* Quoted local part is accepted.
* Space is accepted.
* Symbols have special meaning in some MTAs is rejected.
    - Including (but not limited to): `%`, `|`, `!`, `#`, `$`, `*`, `/`, `\`

## Domain Part

* Option to accept IP literals.
    - Hybrid address (IPv4-mapped IPv6 address) is not accept.
