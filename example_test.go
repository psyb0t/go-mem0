package mem0_test

import (
	"context"
	"net/http"
	"os"

	"github.com/psyb0t/go-mem0"
)

//nolint:testableexamples // Needs real credentials, compile check only.
func Example() {
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
