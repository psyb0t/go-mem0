# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-17

- Added `mem0`, a REST client generated with oapi-codegen from the Mem0 Platform API document (spec version v1, OpenAPI 3.0.1), covering all 47 paths and 68 operations: memories, entities, events, exports, organizations, projects, webhooks, profiles and Dream operations.
- Unformatted number fields are `float64` instead of oapi-codegen's default `float32`.
- `make generate` downloads the current spec from https://docs.mem0.ai/openapi.json and regenerates the complete client without patches. The spec is not committed.
- Added the `go-mem0` agent skill and Claude Code and Codex plugin manifests.
