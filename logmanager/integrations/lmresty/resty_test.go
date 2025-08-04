package lmresty_test

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/internal/test/testdata"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

type TestSuite struct {
	client *resty.Client
	server *httptest.Server
}

func (suite *TestSuite) SetupServer(statusCode int) {
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	}))
}

func (suite *TestSuite) TearDownServer() {
	if suite.server != nil {
		suite.server.Close()
	}
}

func (suite *TestSuite) SetupRestyClient() {
	suite.client = resty.New()
	if suite.server != nil {
		suite.client.SetBaseURL(suite.server.URL)
	}
}

func TestNewTxn_Suite(t *testing.T) {
	suite := &TestSuite{}

	t.Run("Nil response", func(t *testing.T) {
		ctx := context.Background()

		// Set up the mock server and client
		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		// Perform the POST request
		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/mock-endpoint")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		txn := lmresty.NewTxn(nil)
		defer txn.End()
		assert.Nil(t, txn)
	})

	t.Run("Nil transaction in context", func(t *testing.T) {
		ctx := context.Background()

		// Set up the mock server and client
		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		// Perform the POST request
		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/mock-endpoint")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		txn := lmresty.NewTxn(resp)
		defer txn.End()
		assert.Nil(t, txn)
	})

	t.Run("Valid response and transaction in context", func(t *testing.T) {
		ctx := testdata.NewTx("a", "b").ToContext(context.Background())

		// Set up the mock server and client
		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		// Perform the POST request
		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/mock-endpoint")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		txn := lmresty.NewTxn(resp)
		defer txn.End()
		assert.NotNil(t, txn)
	})

	t.Run("Valid response with TestableApplication and field assertions", func(t *testing.T) {
		app := logmanager.NewTestableApplication()

		// Create a transaction using the app's Application StartHttp method
		txnFromApp := app.Application.StartHttp("test-trace-id", "POST /mock-endpoint")
		ctx := txnFromApp.ToContext(context.Background())

		// Reset logged entries before test
		app.ResetLoggedEntries()

		// Set up the mock server and client
		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		// Perform the POST request
		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/mock-endpoint")

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		txn := lmresty.NewTxn(resp)
		assert.NotNil(t, txn)
		txn.End()

		// Verify transaction was logged
		assert.Equal(t, 1, app.CountLoggedEntries(), "Should have exactly one logged entry")
	})
}

// Test suite for masking functionality
func TestMaskingFunctionality(t *testing.T) {
	suite := &TestSuite{}

	t.Run("NewTxnWithMasking - Password field masking", func(t *testing.T) {
		app := logmanager.NewTestableApplication()
		txnFromApp := app.Application.StartHttp("test-trace-id", "POST /login")
		ctx := txnFromApp.ToContext(context.Background())
		app.ResetLoggedEntries()

		// Setup mock server to return sensitive data
		suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"user": "john", "password": "secret123", "token": "abc123"}`))
		}))
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		// Perform request with sensitive data
		resp, err := suite.client.R().
			SetContext(ctx).
			SetBody(`{"username": "john", "password": "mypassword"}`).
			Post("/login")

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		// Create a transaction with password masking
		maskingConfigs := []logmanager.MaskingConfig{
			{
				FieldPattern: "password",
				Type:         logmanager.FullMask,
			},
			{
				FieldPattern: "token",
				Type:         logmanager.FullMask,
			},
		}

		txn := lmresty.NewTxnWithMasking(resp, maskingConfigs)
		assert.NotNil(t, txn)
		txn.End()

		// Verify masking was applied
		assert.Equal(t, 1, app.CountLoggedEntries())
	})

	t.Run("NewTxnWithConfig - CombinedMasking", func(t *testing.T) {
		app := logmanager.NewTestableApplication()
		txnFromApp := app.Application.StartHttp("test-trace-id", "POST /profile")
		ctx := txnFromApp.ToContext(context.Background())
		app.ResetLoggedEntries()

		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/profile")

		assert.NoError(t, err)

		// Test combined masking with config
		maskingConfigs := []logmanager.MaskingConfig{
			{
				FieldPattern: "email",
				Type:         logmanager.PartialMask,
				ShowFirst:    3,
				ShowLast:     5,
			},
			{
				FieldPattern: "ssn",
				Type:         logmanager.PartialMask,
				ShowFirst:    3,
				ShowLast:     4,
			},
		}

		txn := lmresty.NewTxnWithConfig(resp, maskingConfigs)
		assert.NotNil(t, txn)
		txn.End()

		assert.Equal(t, 1, app.CountLoggedEntries())
	})

	t.Run("Convenience functions", func(t *testing.T) {
		app := logmanager.NewTestableApplication()

		// Test password masking convenience function
		t.Run("NewTxnWithPasswordMasking", func(t *testing.T) {
			txnFromApp := app.Application.StartHttp("test-trace-id", "POST /auth")
			ctx := txnFromApp.ToContext(context.Background())
			app.ResetLoggedEntries()

			suite.SetupServer(http.StatusOK)
			defer suite.TearDownServer()
			suite.SetupRestyClient()

			resp, err := suite.client.R().
				SetContext(ctx).
				Post("/auth")

			assert.NoError(t, err)

			txn := lmresty.NewTxnWithPasswordMasking(resp)
			assert.NotNil(t, txn)
			txn.End()

			assert.Equal(t, 1, app.CountLoggedEntries())
		})

		// Test email masking convenience function
		t.Run("NewTxnWithEmailMasking", func(t *testing.T) {
			txnFromApp := app.Application.StartHttp("test-trace-id", "POST /users")
			ctx := txnFromApp.ToContext(context.Background())
			app.ResetLoggedEntries()

			suite.SetupServer(http.StatusOK)
			defer suite.TearDownServer()
			suite.SetupRestyClient()

			resp, err := suite.client.R().
				SetContext(ctx).
				Post("/users")

			assert.NoError(t, err)

			txn := lmresty.NewTxnWithEmailMasking(resp)
			assert.NotNil(t, txn)
			txn.End()

			assert.Equal(t, 1, app.CountLoggedEntries())
		})

		// Test credit card masking convenience function
		t.Run("NewTxnWithCreditCardMasking", func(t *testing.T) {
			txnFromApp := app.Application.StartHttp("test-trace-id", "POST /payment")
			ctx := txnFromApp.ToContext(context.Background())
			app.ResetLoggedEntries()

			suite.SetupServer(http.StatusOK)
			defer suite.TearDownServer()
			suite.SetupRestyClient()

			resp, err := suite.client.R().
				SetContext(ctx).
				Post("/payment")

			assert.NoError(t, err)

			txn := lmresty.NewTxnWithCreditCardMasking(resp)
			assert.NotNil(t, txn)
			txn.End()

			assert.Equal(t, 1, app.CountLoggedEntries())
		})
	})
}

func TestDirectMaskingUsage(t *testing.T) {
	suite := &TestSuite{}

	t.Run("Direct usage with nil configs", func(t *testing.T) {
		app := logmanager.NewTestableApplication()
		txnFromApp := app.Application.StartHttp("test-trace-id", "POST /test")
		ctx := txnFromApp.ToContext(context.Background())
		app.ResetLoggedEntries()

		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/test")

		assert.NoError(t, err)

		// Test with nil configs (should still work)
		txn := lmresty.NewTxnWithConfig(resp, nil)
		assert.NotNil(t, txn)
		txn.End()

		assert.Equal(t, 1, app.CountLoggedEntries())
	})

	t.Run("Direct usage with empty configs", func(t *testing.T) {
		app := logmanager.NewTestableApplication()
		txnFromApp := app.Application.StartHttp("test-trace-id", "POST /test")
		ctx := txnFromApp.ToContext(context.Background())
		app.ResetLoggedEntries()

		suite.SetupServer(http.StatusOK)
		defer suite.TearDownServer()
		suite.SetupRestyClient()

		resp, err := suite.client.R().
			SetContext(ctx).
			Post("/test")

		assert.NoError(t, err)

		// Test with empty configs
		txn := lmresty.NewTxnWithConfig(resp, []logmanager.MaskingConfig{})
		assert.NotNil(t, txn)
		txn.End()

		assert.Equal(t, 1, app.CountLoggedEntries())
	})
}
