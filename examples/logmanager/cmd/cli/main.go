package main

import (
	"context"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmresty"
	"github.com/go-resty/resty/v2"
)

func main() {
	app := logmanager.NewApplication()
	CliHandler(app)
}

func CliHandler(app *logmanager.Application) {
	t := app.Start("abc", "cli", logmanager.TxnTypeOther)

	ctx := t.ToContext(context.Background())
	defer t.End()

	callApiPost(ctx)
	callApiGet(ctx)
}

func callApiPost(ctx context.Context) {
	r := resty.New()
	resp, err := r.R().SetContext(ctx).SetBody(
		map[string]interface{}{
			"foo":   "bar",
			"value": 100,
		},
	).Post("https://httpbin.org/post")

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Println("Response 1 OK")
}

func callApiGet(ctx context.Context) {
	r := resty.New()
	resp, err := r.R().
		SetQueryParams(map[string]string{
			"foo": "bar",
		}).
		SetContext(ctx).               // Use the passed context
		Get("https://httpbin.org/get") // Replace with the actual API endpoint

	txn := lmresty.NewTxn(resp)
	defer txn.End()

	if err != nil {
		txn.NoticeError(err)
	}

	fmt.Println("Response 2 OK")
}
