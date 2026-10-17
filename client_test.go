package mem0_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/psyb0t/go-mem0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAuthorization = "Token test-key"
	jsonContentType   = "application/json"
	testUserID        = "user-example"
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *mem0.ClientWithResponses {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	auth := mem0.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", testAuthorization)

			return nil
		},
	)
	client, err := mem0.NewClientWithResponses(server.URL, auth)
	require.NoError(t, err)

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()

	w.Header().Set("Content-Type", jsonContentType)
	assert.NoError(t, json.NewEncoder(w).Encode(value))
}

func TestPingRead(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status int
		body   string
	}{
		{"success", http.StatusOK, `{"status":"ok"}`},
		{
			"unauthorized",
			http.StatusUnauthorized,
			`{"detail":"Invalid API key."}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t, func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1/ping/", r.URL.Path)
				assert.Equal(t, testAuthorization,
					r.Header.Get("Authorization"))
				w.Header().Set("Content-Type", jsonContentType)
				w.WriteHeader(tc.status)
				_, err := w.Write([]byte(tc.body))
				assert.NoError(t, err)
			})

			res, err := client.PingReadWithResponse(context.Background())
			require.NoError(t, err)
			require.Equal(t, tc.status, res.StatusCode())

			if tc.status == http.StatusUnauthorized {
				require.NotNil(t, res.JSON401)
				assert.Equal(t, "Invalid API key.", *res.JSON401.Detail)
				assert.Nil(t, res.JSON200)

				return
			}

			require.NotNil(t, res.JSON200)
			assert.Equal(t, "ok", *res.JSON200.Status)
		})
	}
}

func TestMemoriesSearchV3(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v3/memories/search/", r.URL.Path)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		assert.Equal(t, jsonContentType, r.Header.Get("Content-Type"))

		var body mem0.MemoriesSearchV3JSONRequestBody

		if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
			return
		}

		assert.Equal(t, "Preferences?", body.Query)
		assert.Equal(t, map[string]any{"user_id": testUserID},
			body.Filters)
		assert.Equal(t, 0.125, *body.Threshold)

		writeJSON(t, w, map[string]any{
			"results": []map[string]any{{
				"id":              "00000000-0000-0000-0000-000000000001",
				"memory":          "Prefers tea",
				"created_at":      time.Time{},
				"updated_at":      nil,
				"expiration_date": nil,
				"metadata":        nil,
				"score_breakdown": map[string]any{
					"semantic": 0.123456789123,
					"bm25":     0.234567891234,
					"entity":   0.345678912345,
				},
			}},
		})
	})

	threshold := 0.125
	res, err := client.MemoriesSearchV3WithResponse(
		context.Background(),
		mem0.MemoriesSearchV3JSONRequestBody{
			Query:     "Preferences?",
			Filters:   map[string]any{"user_id": testUserID},
			Threshold: &threshold,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	require.Len(t, res.JSON200.Results, 1)

	memory := res.JSON200.Results[0]
	assert.Equal(t, "Prefers tea", memory.Memory)
	assert.Nil(t, memory.UpdatedAt)
	assert.Nil(t, memory.ExpirationDate)
	assert.Nil(t, memory.Metadata)
	require.NotNil(t, memory.ScoreBreakdown)
	assert.Equal(t, 0.123456789123, *memory.ScoreBreakdown.Semantic)
	assert.Equal(t, 0.234567891234, *memory.ScoreBreakdown.Bm25)
	assert.Equal(t, 0.345678912345, *memory.ScoreBreakdown.Entity)
}

func TestPingMalformedJSON(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", jsonContentType)
		_, err := w.Write([]byte(`{"status":}`))
		assert.NoError(t, err)
	})

	res, err := client.PingReadWithResponse(context.Background())

	var syntaxError *json.SyntaxError
	require.ErrorAs(t, err, &syntaxError)
	assert.Nil(t, res)
}

func TestPingCanceled(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("canceled request reached the server")
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := client.PingReadWithResponse(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, res)
}

func TestMemoriesListPagination(t *testing.T) {
	t.Parallel()

	const pageSize = 2

	identities := []string{
		"memory-1", "memory-2", "memory-3",
		"memory-4", "memory-5",
	}
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/memories/", r.URL.Path)
		assert.Equal(t, testAuthorization, r.Header.Get("Authorization"))
		assert.Equal(t, testUserID, r.URL.Query().Get("user_id"))
		assert.Equal(t, strconv.Itoa(pageSize),
			r.URL.Query().Get("page_size"))

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if !assert.NoError(t, err) {
			return
		}

		start := min((page-1)*pageSize, len(identities))
		end := min(start+pageSize, len(identities))
		results := make([]map[string]any, 0, end-start)

		for _, id := range identities[start:end] {
			results = append(results, map[string]any{
				"id": id, "memory": "Prefers tea",
				"created_at": time.Time{}, "updated_at": time.Time{},
			})
		}

		var next *string

		if end < len(identities) {
			path := "/v1/memories/?page=" + strconv.Itoa(page+1)
			next = &path
		}

		writeJSON(t, w, map[string]any{
			"count": len(identities), "next": next,
			"previous": nil, "results": results,
		})
	})

	userID := testUserID
	size := pageSize

	var collected []string

	for _, page := range []int{1, 2, 3, 4, 100} {
		res, err := client.MemoriesListWithResponse(
			context.Background(),
			&mem0.MemoriesListParams{
				UserId: &userID, Page: &page, PageSize: &size,
			},
		)
		require.NoError(t, err)
		require.NotNil(t, res.JSON200)
		body, err := res.JSON200.AsMemoriesList200JSONResponseBody1()
		require.NoError(t, err)
		assert.Equal(t, len(identities), body.Count)

		if page > 3 {
			assert.Empty(t, body.Results)
			assert.Nil(t, body.Next)

			continue
		}

		assert.Len(t, body.Results, min(pageSize, len(identities)-
			(page-1)*pageSize))
		assert.Equal(t, page < 3, body.Next != nil)

		for _, memory := range body.Results {
			collected = append(collected, memory.Id)
		}
	}

	assert.Equal(t, identities, collected)
}
