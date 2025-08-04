package async

import (
	"context"
	"encoding/json"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"net/http"
	"time"
)

type Handler struct {
}

func (h Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	syncProcess(ctx)                                                                      // normal process
	go asyncProcessFirst(logmanager.CloneTransactionToContext(ctx, context.Background())) // async process 1
	go asyncProcessLast(logmanager.CloneTransactionToContext(ctx, context.Background()))  // async process 2

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "ok",
	})
}

func syncProcess(ctx context.Context) {
	txn := logmanager.StartOtherSegment(logmanager.FromContext(ctx), logmanager.OtherSegment{
		Name: "sync-process",
	})
	defer txn.End()
	time.Sleep(1000 * time.Millisecond)
}

func asyncProcessFirst(ctx context.Context) {
	txn := logmanager.StartOtherSegment(logmanager.FromContext(ctx), logmanager.OtherSegment{
		Name: "async-process-1",
	})
	defer txn.End()
	time.Sleep(1000 * time.Millisecond)
}

func asyncProcessLast(ctx context.Context) {
	txn := logmanager.StartOtherSegment(logmanager.FromContext(ctx), logmanager.OtherSegment{
		Name: "async-process-2",
	})
	defer txn.End()
	time.Sleep(1000 * time.Millisecond)
}
