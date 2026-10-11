# Agent guidance

## Consumer API invariants
- Preserve documented return values and consumed-versus-remaining input semantics for UntilConsumer, PrefixConsumer and other consumers.
- This library uses typed options on some consumption APIs. Preserve existing call-site compatibility; where the code already uses type-switched variadic options, add named types and one locally resolved options configuration rather than requiring new parameters everywhere.
- Confirm the *actual* implementation's precedence for repeated options and its policy for unsupported options before documenting or modifying either. Do not introduce a blanket panic policy.
- Test separator matching, escaping, encasing, empty input, missing delimiters and resulting consumed/remaining slices. Include negative and boundary cases and readable examples.
- For pure string parsing, avoid incidental filesystem and process setup. A filesystem abstraction is unnecessary if no file IO occurs.

## Agent handoff
Read README and existing examples, then run relevant Go tests and report exact edge-case gaps and API-compatibility risks in issue #8. Do not change exported signatures merely to simplify testing.
## Engineering reference and follow-up

- Adopt only relevant guidance from https://arran4.github.io/blog/post/2026/055-suggested-coding-standards-for-repositories-and-agents/ and https://arran4.github.io/blog/post/2026/020-optional-dependency-injection-via-type-switched-variadic-args/ .
- This documentation is not evidence of repository-wide compliance. A coding-agent pass must report checked files, observed gaps, tests/CI and existing issue overlap to https://github.com/arran4/go-consume/issues/8 before any broad implementation.
- Do not merge implementation automatically; request review and follow repository-specific instructions.
