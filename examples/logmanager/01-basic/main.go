package main

import (
	"context"
	"fmt"

	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/go-resty/resty/v2"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithAppName("basic-cli"),
	)

	txn := app.Start("demo-cli", "cli", logmanager.TxnTypeOther)
	ctx := txn.ToContext(context.Background())
	defer txn.End()

	makePostRequest(ctx)
	makeGetRequest(ctx)
}

func makePostRequest(ctx context.Context) {
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"foo":   "bar",
			"value": 100,
		}).
		Post("https://httpbin.org/post")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		return
	}

	fmt.Println("POST request completed successfully")
}

func makeGetRequest(ctx context.Context) {
	client := resty.New()
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"foo": "bar",
		}).
		Get("https://httpbin.org/get")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
		return
	}

	fmt.Println("GET request completed successfully")
}