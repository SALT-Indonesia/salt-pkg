package clientmanager_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	method    int
	option    []clientmanager.Option
	isSuccess bool
}

func basicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || len(strings.ReplaceAll(auth, "Bearer ", "")) == 0 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func apiKeyAuth(key, value string, addToQueryParams bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get(key)
			if addToQueryParams {
				auth = r.URL.Query().Get(key)
			}
			if auth != value {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
})

func test(t *testing.T, testCases []testCase, ts *httptest.Server) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	for _, tc := range testCases {
		t.Run(http.StatusText(tc.method), func(t *testing.T) {
			res, err := clientmanager.Call[string](
				ctx,
				ts.URL,
				tc.option...,
			)

			assert.NotNil(t, res)
			assert.NoError(t, err)
			assert.Equal(t, tc.method, res.StatusCode)
			assert.Equal(t, tc.isSuccess, res.IsSuccess())
		})
	}
}

func TestAuthBasic(t *testing.T) {
	user, pass := "user123", "pass123"
	handler := basicAuth(user, pass)(handler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	test(t, []testCase{
		{
			method:    http.StatusOK,
			option:    []clientmanager.Option{clientmanager.WithAuth(clientmanager.AuthBasic(user, pass))},
			isSuccess: true,
		},
		{
			method:    http.StatusUnauthorized,
			isSuccess: false,
		},
	}, ts)
}

func TestAuthBearer(t *testing.T) {
	handler := bearerAuth(handler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	test(t, []testCase{
		{
			method:    http.StatusOK,
			option:    []clientmanager.Option{clientmanager.WithAuth(clientmanager.AuthBearer("mytoken"))},
			isSuccess: true,
		},
		{
			method:    http.StatusUnauthorized,
			isSuccess: false,
		},
	}, ts)
}

func TestAuthAPIKey(t *testing.T) {
	key, value := "api_key", "myapikey"

	t.Run("in header", func(t *testing.T) {
		handler := apiKeyAuth(key, value, false)(handler)
		ts := httptest.NewServer(handler)
		defer ts.Close()

		test(t, []testCase{
			{
				method:    http.StatusOK,
				option:    []clientmanager.Option{clientmanager.WithAuth(clientmanager.AuthAPIKey(key, value, false))},
				isSuccess: true,
			},
			{
				method:    http.StatusUnauthorized,
				isSuccess: false,
			},
		}, ts)
	})

	t.Run("in query param", func(t *testing.T) {
		handler := apiKeyAuth(key, value, true)(handler)
		ts := httptest.NewServer(handler)
		defer ts.Close()

		test(t, []testCase{
			{
				method:    http.StatusOK,
				option:    []clientmanager.Option{clientmanager.WithAuth(clientmanager.AuthAPIKey(key, value, true))},
				isSuccess: true,
			},
			{
				method:    http.StatusUnauthorized,
				isSuccess: false,
			},
		}, ts)
	})
}
