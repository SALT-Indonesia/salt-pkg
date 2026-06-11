package clientmanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var (
	testHandlerFunc = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
)

// withClient is a test helper function to set a custom HTTP client
func withClient(client *http.Client) Option {
	return func(co *callOptions) {
		co.client = client
	}
}

func testClient(ctx context.Context, client *http.Client, tsURL string) (time.Duration, error) {
	requests, eg, start := 100, new(errgroup.Group), time.Now()
	for range requests {
		eg.Go(func() error {
			if _, err := Call[any](logmanager.CloneTransactionToContext(ctx, context.Background()), tsURL, withClient(client)); err != nil {
				return err
			}
			return nil
		})
	}

	err := eg.Wait()

	return time.Since(start), err
}

func TestClient(t *testing.T) {
	ts := httptest.NewServer(testHandlerFunc)
	defer ts.Close()

	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	txn2 := app.Start("test-2", "cli-2", logmanager.TxnTypeOther)
	ctx2 := txn2.ToContext(context.Background())
	defer txn2.End()

	defaultDuration, defaultErr := testClient(ctx, http.DefaultClient, ts.URL)
	tunedDuration, tunedErr := testClient(ctx2, client, ts.URL)

	assert.NoError(t, defaultErr)
	assert.NoError(t, tunedErr)
	// Both clients should complete within reasonable time
	assert.Less(t, defaultDuration, 10*time.Second)
	assert.Less(t, tunedDuration, 10*time.Second)
}

func TestTransportOptions(t *testing.T) {
	t.Run("WithResponseHeaderTimeout sets transport field", func(t *testing.T) {
		opts := &callOptions{client: newClient()}
		WithResponseHeaderTimeout(30 * time.Second)(opts)

		tr, ok := opts.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, 30*time.Second, tr.ResponseHeaderTimeout)
	})

	t.Run("WithTimeout raises ResponseHeaderTimeout when shorter", func(t *testing.T) {
		opts := &callOptions{client: newClient()}
		// Default ResponseHeaderTimeout is 5s, setting a 120s timeout should raise it
		WithTimeout(120 * time.Second)(opts)

		tr, ok := opts.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, 120*time.Second, tr.ResponseHeaderTimeout)
		assert.Equal(t, 120*time.Second, opts.client.Timeout)
	})

	t.Run("WithTimeout does not lower ResponseHeaderTimeout", func(t *testing.T) {
		opts := &callOptions{client: newClient()}
		// First set a generous header timeout
		WithResponseHeaderTimeout(60 * time.Second)(opts)
		// Then set a shorter overall timeout — header timeout should stay at 60s
		WithTimeout(10 * time.Second)(opts)

		tr, ok := opts.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, 60*time.Second, tr.ResponseHeaderTimeout)
		assert.Equal(t, 10*time.Second, opts.client.Timeout)
	})

	t.Run("WithTimeout does not change ResponseHeaderTimeout when timeout is below default", func(t *testing.T) {
		opts := &callOptions{client: newClient()}
		// Setting a 3s timeout should not affect the 5s default header timeout
		WithTimeout(3 * time.Second)(opts)

		tr, ok := opts.client.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Equal(t, 5*time.Second, tr.ResponseHeaderTimeout)
	})
}

func BenchmarkClients(b *testing.B) {
	ts := httptest.NewServer(testHandlerFunc)
	defer ts.Close()

	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	clients := map[string]*http.Client{
		"default": http.DefaultClient,
		"tuned":   client,
	}

	for name, client := range clients {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := testClient(ctx, client, ts.URL); err != nil {
					b.Error(err)
				}
			}
		})
	}
}
