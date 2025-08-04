package logmanager

type TxnType string

const (
	TxnTypeHttp     TxnType = "http"
	TxnTypeApi      TxnType = "api"
	TxnTypeDatabase TxnType = "database"
	TxnTypeConsumer TxnType = "consumer"
	TxnTypeCron     TxnType = "cron"
	TxnTypeGrpc     TxnType = "grpc"
	TxnTypeOther    TxnType = "other"
)
