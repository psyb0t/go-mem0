# go-mem0

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-mem0.svg)](https://pkg.go.dev/github.com/psyb0t/go-mem0)
[![CI](https://github.com/psyb0t/go-mem0/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-mem0/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-mem0/badges/version.svg)](https://github.com/psyb0t/go-mem0/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-mem0/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-mem0/badges/importers.svg)](https://github.com/psyb0t/go-mem0/blob/badges/importers.md)

Unofficial Go client for the [Mem0 Platform API](https://docs.mem0.ai), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from the OpenAPI document Mem0 publishes. I am not affiliated with Mem0.

The package is `mem0`, in the repository root. Everything in `client.gen.go` is generated. Don't edit it by hand, regenerate it.

## Install

```sh
go get github.com/psyb0t/go-mem0
```

## Usage

Every endpoint in the spec has a typed method on `ClientWithResponses`. Responses decode into `JSON200` and the other status fields.

```go
package main

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-mem0"
)

func main() {
	auth := mem0.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(
				"Authorization",
				"Token "+os.Getenv("MEM0_API_KEY"),
			)

			return nil
		},
	)

	client, err := mem0.NewClientWithResponses(
		"https://api.mem0.ai",
		auth,
	)
	if err != nil {
		panic(err)
	}

	res, err := client.PingReadWithResponse(context.Background())
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
```

The same code is in `example_test.go`, so `make test` checks that it compiles.

Method names come from the spec's operation IDs. The client includes memory creation, retrieval, search, updates and deletion, entity management, exports, organizations, projects, webhooks, profiles and Dream operations. Both current and deprecated endpoints are generated.

## Auth

Mem0 uses the spec's `ApiKeyAuth` scheme, an API key in the `Authorization` header with a `Token ` prefix. The request editor above reads the key from `MEM0_API_KEY` and sets `Authorization: Token <key>`. Never log this header or put a key in source code.

The Platform API base URL is `https://api.mem0.ai`. Use `WithHTTPClient` to supply an HTTP client with your own timeout.

## Regenerating

The spec is not committed. `make generate` downloads the current OpenAPI document into the gitignored `api/` directory and regenerates the client:

```sh
make generate
```

`gen.go` declares the steps, so `make generate` is just `go generate ./...`:

- `scripts/fetch-spec.sh` downloads `https://docs.mem0.ai/openapi.json` to `api/openapi.json`.
- oapi-codegen runs with `config.yaml` over the complete OpenAPI 3.0.1 document. No spec patch is needed.
- The config maps unformatted `number` fields to `float64`. Fields explicitly formatted as `float` keep the generator's `float32` type.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-mem0) teaches a coding agent how to use go-mem0 when it writes Go. It works in agents that read `.agents/skills/`. The commands below apply after the repository is published and registered in the marketplace.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-mem0@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-mem0@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The release workflow is configured to publish the skill to ClawHub. After publication:

```bash
openclaw skills install @psyb0t/go-mem0
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the spec and regenerate
```

## License

[MIT](LICENSE) for the handwritten repository files. Mem0's source specification declares Apache 2.0; generated code derives from that specification. The downloaded specification is not redistributed here.
