# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.
- The skill's frontmatter is valid YAML again; an unquoted colon in the description stopped it from parsing.

## v0.1.0, 2026-10-17

- Added `mem0`, a REST client generated with oapi-codegen from the Mem0 Platform API document (spec version v1, OpenAPI 3.0.1), covering all 47 paths and 68 operations: memories, entities, events, exports, organizations, projects, webhooks, profiles and Dream operations.
- Unformatted number fields are `float64` instead of oapi-codegen's default `float32`.
- `make generate` downloads the current spec from https://docs.mem0.ai/openapi.json and regenerates the complete client without patches. The spec is not committed.
- Added the `go-mem0` agent skill and Claude Code and Codex plugin manifests.
