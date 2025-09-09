package clientmanager_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
	"github.com/SALT-Indonesia/salt-pkg/clientmanager/examples/dummyjson/product"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

var (
	req = &product.Request{
		Title: "Essence Mascara Lash Princess",
		Price: 9.99,
		Stock: 5,
	}
)

func TestCallGET(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("valid JSON", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"products": [
					{
						"id": 1,
						"title": "Essence Mascara Lash Princess",
						"price": 9.99,
						"stock": 5
					},
					{
						"id": 2,
						"title": "ThinkPad T14",
						"price": 1600,
						"stock": 9
					},
					{
						"id": 3,
						"title": "MacBook Pro",
						"price": 2000,
						"stock": 9
					}
				],
				"total": 3,
				"skip": 0,
				"limit": 10
			}`))
		}))
		defer ts.Close()

		res, err := clientmanager.Call[product.Response](
			ctx,
			"",
			clientmanager.WithHost(ts.URL),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
		assert.Len(t, res.Body.Products, 3)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"products": [
					{
						"id": 1,
						"title": "Essence Mascara Lash Princess",
						"price": 9.99,
						"stock": 5,
					},
					{
						"id": 2,
						"title": "ThinkPad T14",
						"price": 1600,
						"stock": 9,
					},
					{
						"id": 3,
						"title": "MacBook Pro",
						"price": 2000,
						"stock": 9,
					}
				],
				"total": 3,
				"skip": 0,
				"limit": 10,
			}`))
		}))
		defer ts.Close()

		_, err := clientmanager.Call[product.Response](
			ctx,
			"",
			clientmanager.WithHost(ts.URL),
		)

		assert.Error(t, err)
	})

	t.Run("with insecure", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"products": [
					{
						"id": 1,
						"title": "Essence Mascara Lash Princess",
						"price": 9.99,
						"stock": 5
					}
				],
				"total": 1,
				"skip": 0,
				"limit": 10
			}`))
		}))
		defer ts.Close()

		res, err := clientmanager.Call[product.Response](
			ctx,
			"",
			clientmanager.WithHost(ts.URL),
			clientmanager.WithInsecure(),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
	})

	t.Run("with url values", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"products": [
					{
						"id": 1,
						"title": "Annibale Colombo Bed",
						"price": 1899.99
					}
				],
				"total": 1,
				"skip": 0,
				"limit": 10
			}`))
		}))
		defer ts.Close()

		res, err := clientmanager.Call[product.Response](
			ctx,
			"",
			clientmanager.WithURLValues(url.Values{
				"limit":  {"10"},
				"skip":   {"10"},
				"select": {"title,price"},
			}),
			clientmanager.WithHost(ts.URL),
		)

		assert.NotNil(t, res)
		assert.NoError(t, err)
	})

	t.Run("with timeout", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"products": [
					{
						"id": 1,
						"title": "Annibale Colombo Bed",
						"price": 1899.99
					}
				],
				"total": 1,
				"skip": 0,
				"limit": 10
			}`))
		}))
		defer ts.Close()

		res, err := clientmanager.Call[product.Response](
			ctx,
			"",
			clientmanager.WithHost(ts.URL),
			clientmanager.WithTimeout(time.Second),
		)

		assert.NotNil(t, res)
		assert.NoError(t, err)
	})

	t.Run("without transaction context", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		res, err := clientmanager.Call[product.Response](
			context.Background(),
			ts.URL,
		)

		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestCallPOST(t *testing.T) {
	testCases := []struct {
		name       string
		server     *httptest.Server
		host       string
		headers    http.Header
		request    *product.Request
		statusCode int
		isSuccess  bool
	}{
		{
			name: "with request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"id": 1,
					"title": "Essence Mascara Lash Princess",
					"price": 9.99,
					"stock": 5
				}`))
			})),
			request:    req,
			statusCode: http.StatusOK,
			isSuccess:  true,
		},
		{
			name: "without request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			statusCode: http.StatusBadRequest,
		},
		{
			name: "empty request",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
			request:    &product.Request{},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "with custom headers",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{
					"id": 1,
					"title": "Essence Mascara Lash Princess",
					"price": 9.99,
					"stock": 5
				}`))
			})),
			headers: map[string][]string{
				"X-Trace-ID": {"abc123"},
			},
			request:    req,
			statusCode: http.StatusOK,
			isSuccess:  true,
		},
		{
			name: "invalid URL",
			host: "://notfound",
		},
		{
			name: "host not found",
			host: "http://notfound",
		},
	}

	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			host := tc.host
			if tc.server != nil {
				defer tc.server.Close()
				host = tc.server.URL
			}

			res, err := clientmanager.Call[any](
				ctx,
				"",
				clientmanager.WithMethod(http.MethodPost),
				clientmanager.WithRequestBody(tc.request),
				clientmanager.WithHost(host),
				clientmanager.WithHeaders(tc.headers),
			)

			if res != nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.statusCode, res.StatusCode)
				assert.Equal(t, tc.isSuccess, res.IsSuccess())
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestClassicForm(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	req := &request{
		Username: "alice",
		Password: "s3cr3t",
	}
	res, err := clientmanager.Call[any](
		ctx,
		"",
		clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
		clientmanager.WithHost(ts.URL),
		clientmanager.WithFormURLEncoded(),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, res.IsSuccess())
}

func TestFileUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	f, _ := os.CreateTemp("", "")
	defer func() {
		_ = os.Remove(f.Name())
	}()

	options := []clientmanager.Option{
		clientmanager.WithMethod(http.MethodPost),
		clientmanager.WithHost(ts.URL),
		clientmanager.WithFiles(map[string]string{
			"image": f.Name(),
		}),
	}

	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("with a file", func(t *testing.T) {
		res, err := clientmanager.Call[any](ctx, "", options...)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
	})

	t.Run("with a missing file", func(t *testing.T) {
		_, err := clientmanager.Call[any](ctx, "", append(
			options,
			clientmanager.WithFiles(map[string]string{
				"image": "missingimage.jpg",
			}),
		)...)

		assert.Error(t, err)
	})

	t.Run("with a request body", func(t *testing.T) {
		req := &product.Request{
			Title: "Essence Mascara Lash Princess",
			Price: 9.99,
			Stock: 5,
		}
		res, err := clientmanager.Call[any](
			ctx,
			"",
			append(
				options,
				clientmanager.WithRequestBody(req),
			)...,
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
	})

	t.Run("with JSON sub-request body", func(t *testing.T) {
		type request struct {
			Product product.Request `json:"product"`
		}

		req := &request{
			Product: product.Request{
				Title: "Essence Mascara Lash Princess",
				Price: 9.99,
				Stock: 5,
			},
		}
		res, err := clientmanager.Call[any](
			ctx,
			"",
			clientmanager.WithMethod(http.MethodPost),
			clientmanager.WithHost(ts.URL),
			clientmanager.WithFiles(map[string]string{
				"image": f.Name(),
			}),
			clientmanager.WithRequestBody(req),
		)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
	})
}

func TestString(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	body := "hello world"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	res, err := clientmanager.Call[string](
		ctx,
		"",
		clientmanager.WithHost(ts.URL),
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.True(t, res.IsSuccess())
	assert.Equal(t, body, res.Body)
}

func TestXML(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	type RSS struct {
		Channel struct {
			Title string `xml:"title"`
		} `xml:"channel"`
	}

	t.Run("valid XML", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "text/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" version="2.0">
				<channel>
					<title><![CDATA[BBC News]]></title>
				</channel>
			</rss>`))
		}))
		defer ts.Close()

		res, err := clientmanager.Call[RSS](ctx, "", clientmanager.WithHost(ts.URL))

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.True(t, res.IsSuccess())
		assert.Equal(t, "BBC News", res.Body.Channel.Title)
	})

	t.Run("invalid XML", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "text/xml")
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
				<rss xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:content="http://purl.org/rss/1.0/modules/content/" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" version="2.0">
				<channel>
					<title><![CDATA[BBC News]]><title>
				</channel>
			</rss>`))
		}))
		defer ts.Close()

		_, err := clientmanager.Call[RSS](ctx, "", clientmanager.WithHost(ts.URL))

		assert.Error(t, err)
	})
}

func TestValidation(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	type request struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	req := &request{
		Username: "alice",
	}
	res, err := clientmanager.Call[any](
		ctx,
		"",
		clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
		clientmanager.WithHost(ts.URL),
	)

	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestProxy(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	t.Run("valid proxy", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}))
		defer ts.Close()

		proxyTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			targetURL, _ := url.Parse(ts.URL)
			req, _ := http.NewRequest(r.Method, targetURL.String(), r.Body)
			req.Header = r.Header

			resp, _ := http.DefaultTransport.RoundTrip(req)
			defer func() {
				_ = resp.Body.Close()
			}()

			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		}))
		defer proxyTS.Close()

		proxy, err := clientmanager.WithProxy(proxyTS.URL)

		assert.NoError(t, err)

		clientManager := clientmanager.New[string]() // we create a new client manager to create a different HTTP client

		res, err := clientManager.Call(
			ctx,
			ts.URL,
			proxy,
		)

		assert.NotNil(t, res)
		assert.Equal(t, "OK", res.Body)
		assert.NoError(t, err)
	})

	t.Run("invalid proxy", func(t *testing.T) {
		proxy, err := clientmanager.WithProxy("://notfound")

		assert.Nil(t, proxy)
		assert.Error(t, err)
	})
}

func TestCertificates(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	res, err := clientmanager.Call[any](
		ctx,
		ts.URL,
		clientmanager.WithCertificates(tls.Certificate{}), // empty certificate as dummy
	)

	assert.NotNil(t, res)
	assert.NoError(t, err)
}

func TestAuth(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Run("basic auth", func(t *testing.T) {
		res, err := clientmanager.Call[any](
			ctx,
			ts.URL,
			clientmanager.WithAuth(clientmanager.AuthBasic("user123", "pass123")),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("bearer token", func(t *testing.T) {
		res, err := clientmanager.Call[any](
			ctx,
			ts.URL,
			clientmanager.WithAuth(clientmanager.AuthBearer("pass123")),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("api key", func(t *testing.T) {
		t.Run("in header", func(t *testing.T) {
			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				clientmanager.WithAuth(clientmanager.AuthAPIKey("foo", "bar", false)),
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("in body", func(t *testing.T) {
			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				clientmanager.WithAuth(clientmanager.AuthAPIKey("foo", "bar", true)),
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})
	})

	t.Run("JWT bearer", func(t *testing.T) {
		t.Run("valid JWT", func(t *testing.T) {
			auth := clientmanager.AuthJWT(
				"mysecretkey",
				jwt.SigningMethodHS256,
				clientmanager.AuthJWTClaims{
					Sub: "myusername",
					Iss: "myissuer",
					Aud: "myaudience",
					Nbf: time.Now(),
					Exp: time.Now().Add(time.Hour),
					Jti: clientmanager.AuthJWTClaimsJWTID{
						Generate: true,
					},
					Extra: map[string]any{
						"name": "John Doe",
					},
				},
			)

			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				clientmanager.WithAuth(auth),
			)

			assert.NotNil(t, res)
			assert.NoError(t, err)
		})

		t.Run("with custom JWT ID", func(t *testing.T) {
			auth := clientmanager.AuthJWT(
				"mysecretkey",
				jwt.SigningMethodHS256,
				clientmanager.AuthJWTClaims{
					Sub: "myusername",
					Iss: "myissuer",
					Aud: "myaudience",
					Nbf: time.Now(),
					Exp: time.Now().Add(time.Hour),
					Jti: clientmanager.AuthJWTClaimsJWTID{
						Generate: true,
						Custom:   "123",
					},
					Extra: map[string]any{
						"name": "John Doe",
					},
				},
			)

			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				clientmanager.WithAuth(auth),
			)

			assert.NotNil(t, res)
			assert.NoError(t, err)
		})
	})

	t.Run("digest auth", func(t *testing.T) {
		t.Run("global", func(t *testing.T) {
			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				clientmanager.WithAuthDigest("user123", "pass123"),
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("new digest transport", func(t *testing.T) {
			clientManager := clientmanager.New[any]()
			res, err := clientManager.Call(
				ctx,
				ts.URL,
				clientmanager.WithAuthDigest("user123", "pass123"),
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("replace digest auth to proxy", func(t *testing.T) {
			proxyTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				targetURL, _ := url.Parse(ts.URL)
				req, _ := http.NewRequest(r.Method, targetURL.String(), r.Body)
				req.Header = r.Header

				resp, _ := http.DefaultTransport.RoundTrip(req)
				defer func() {
					_ = resp.Body.Close()
				}()

				w.WriteHeader(resp.StatusCode)
				_, _ = io.Copy(w, resp.Body)
			}))
			defer proxyTS.Close()

			proxy, _ := clientmanager.WithProxy(proxyTS.URL)
			clientManager := clientmanager.New[any]()
			res, err := clientManager.Call(
				ctx,
				ts.URL,
				clientmanager.WithAuthDigest("user123", "pass123"), // put auth digest first
				proxy, // then replace it with proxy
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("replace digest auth to insecure", func(t *testing.T) {
			clientManager := clientmanager.New[any]()
			res, err := clientManager.Call(
				ctx,
				ts.URL,
				clientmanager.WithAuthDigest("user123", "pass123"), // put auth digest first
				clientmanager.WithInsecure(),                       // then replace it with insecure
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})
	})

	t.Run("oauth1", func(t *testing.T) {
		oauth1 := clientmanager.WithOAuth1(clientmanager.OAuth1Parameters{
			ConsumerKey:    "your_consumer_key",
			ConsumerSecret: "your_consumer_secret",
			AccessToken:    "your_access_token",
			TokenSecret:    "your_access_token_secret",
		})
		res, err := clientmanager.Call[any](
			ctx,
			ts.URL,
			oauth1,
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("oauth2", func(t *testing.T) {
		endpointTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "dummy-token",
				"token_type":   "bearer",
				"expires_in":   3600,
			})
		}))
		defer endpointTS.Close()

		redirectTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer redirectTS.Close()

		t.Run("with access token", func(t *testing.T) {
			oauth2, err := clientmanager.WithOAuth2(clientmanager.OAuth2Parameters[string]{
				Auth: "mytoken",
			})

			assert.NotNil(t, oauth2)
			assert.NoError(t, err)

			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				oauth2,
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("without access token", func(t *testing.T) {
			oauth2, err := clientmanager.WithOAuth2(clientmanager.OAuth2Parameters[string]{})

			assert.NotNil(t, oauth2)
			assert.NoError(t, err)
		})

		t.Run("with config", func(t *testing.T) {
			oauth2, err := clientmanager.WithOAuth2(clientmanager.OAuth2Parameters[*oauth2.Config]{
				Auth: &oauth2.Config{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					Scopes:       []string{"user"},
					Endpoint: oauth2.Endpoint{
						TokenURL: endpointTS.URL,
					},
					RedirectURL: redirectTS.URL,
				},
				CodeFromCallback: "code-from-callback",
			})

			assert.NotNil(t, oauth2)
			assert.NoError(t, err)

			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				oauth2,
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("with unaccepted endpoint", func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			oauth2, err := clientmanager.WithOAuth2(clientmanager.OAuth2Parameters[*oauth2.Config]{
				Auth: &oauth2.Config{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					Scopes:       []string{"user"},
					Endpoint: oauth2.Endpoint{
						TokenURL: ts.URL,
					},
					RedirectURL: redirectTS.URL,
				},
				CodeFromCallback: "code-from-callback",
			})

			assert.Nil(t, oauth2)
			assert.Error(t, err)
		})

		t.Run("without code from callback", func(t *testing.T) {
			oauth2, err := clientmanager.WithOAuth2(clientmanager.OAuth2Parameters[*oauth2.Config]{
				Auth: &oauth2.Config{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					Scopes:       []string{"user"},
					Endpoint: oauth2.Endpoint{
						TokenURL: endpointTS.URL,
					},
					RedirectURL: redirectTS.URL,
				},
			})

			assert.NotNil(t, oauth2)
			assert.NoError(t, err)

			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				oauth2,
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})
	})

	t.Run("hawk", func(t *testing.T) {
		res, err := clientmanager.Call[any](
			ctx,
			ts.URL,
			clientmanager.WithAuth(clientmanager.AuthHawk("myid", "mykey", nil)),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("AWS", func(t *testing.T) {
		auth := clientmanager.WithAuth(clientmanager.AuthAWS(clientmanager.AWSParameters{
			Key:     "mykey",
			Secret:  "mysecretkey",
			Service: "myservice",
			Region:  "ap-southeast-3",
		}))

		t.Run("without request body", func(t *testing.T) {
			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				auth,
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("with request body", func(t *testing.T) {
			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				auth,
				clientmanager.WithMethod(http.MethodPost),
				clientmanager.WithRequestBody(req),
			)

			assert.NotNil(t, res)
			assert.True(t, res.IsSuccess())
			assert.NoError(t, err)
		})

		t.Run("without secret", func(t *testing.T) {
			res, err := clientmanager.Call[any](
				ctx,
				ts.URL,
				clientmanager.WithAuth(clientmanager.AuthAWS(clientmanager.AWSParameters{
					Key:     "mykey",
					Service: "myservice",
					Region:  "ap-southeast-3",
				})),
			)

			assert.Nil(t, res)
			assert.Error(t, err)
		})
	})

	t.Run("NTLM", func(t *testing.T) {
		res, err := clientmanager.Call[any](
			ctx,
			ts.URL,
			clientmanager.WithAuthNTLM(clientmanager.AuthBasic("user123", "pass123")),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("ESB", func(t *testing.T) {
		res, err := clientmanager.Call[any](
			ctx,
			ts.URL,
			clientmanager.WithAuth(clientmanager.AuthESB("user123", "pass123")),
			clientmanager.WithMethod(http.MethodPost),
			clientmanager.WithRequestBody(req),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})
}

func TestWithConnection(t *testing.T) {
	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	clientManager := clientmanager.New[any]()

	t.Run("with connection limit", func(t *testing.T) {
		res, err := clientManager.Call(
			ctx,
			ts.URL,
			clientmanager.WithConnectionLimit(1000, 1000, 100),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("with idle connection timeout", func(t *testing.T) {
		res, err := clientManager.Call(
			ctx,
			ts.URL,
			clientmanager.WithIdleConnTimeout(time.Minute),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("with TLS handshake timeout", func(t *testing.T) {
		res, err := clientManager.Call(
			ctx,
			ts.URL,
			clientmanager.WithTLSHandshakeTimeout(5*time.Second),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("with expect continue timeout", func(t *testing.T) {
		res, err := clientManager.Call(
			ctx,
			ts.URL,
			clientmanager.WithExpectContinueTimeout(5*time.Second),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})

	t.Run("with dial context", func(t *testing.T) {
		res, err := clientManager.Call(
			ctx,
			ts.URL,
			clientmanager.WithDialContext(10*time.Second, time.Minute),
		)

		assert.NotNil(t, res)
		assert.True(t, res.IsSuccess())
		assert.NoError(t, err)
	})
}
