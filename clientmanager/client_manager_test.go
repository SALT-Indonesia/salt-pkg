package clientmanager_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SALT-Indonesia/salt-pkg/clientmanager"
	"github.com/SALT-Indonesia/salt-pkg/clientmanager/examples/dummyjson/product"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
					"id": 1,
					"title": "Essence Mascara Lash Princess",
					"price": 9.99,
					"stock": 5
				}`))
	}))
	defer ts.Close()

	app := logmanager.NewApplication()
	txn := app.Start("test", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	clientManager := clientmanager.New[product.Product](
		clientmanager.WithHost(ts.URL),
	)
	res, err := clientManager.Call(ctx, "")

	assert.NotNil(t, res)
	assert.NoError(t, err)
}
