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
	assert.Less(t, tunedDuration, defaultDuration) //todo check this
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
