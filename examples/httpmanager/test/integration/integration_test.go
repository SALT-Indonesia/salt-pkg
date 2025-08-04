package integration_test

import (
	"encoding/json"
	"examples/httpmanager/internal/delivery/home"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", home.NewHandler()).Methods("GET")
	//r.Handle("/me", profile.NewHandler()).Methods("GET")
	return r
}

func TestGetEndpoint(t *testing.T) {
	r := setupRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name     string
		url      string
		status   int
		expected map[string]interface{}
	}{
		{
			name:   "get home page",
			url:    "/",
			status: http.StatusOK,
			expected: map[string]interface{}{
				"message": "ok",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, _ := http.Get(ts.URL + tc.url)
			defer resp.Body.Close()
			require.Equal(t, tc.status, resp.StatusCode)

			var body map[string]interface{}
			err := json.NewDecoder(resp.Body).Decode(&body)
			require.NoError(t, err, "failed to decode response body")
			require.Equal(t, tc.expected, body)
		})
	}
}
