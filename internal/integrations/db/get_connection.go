package db

import (
	"context"
	"github.com/go-pg/pg/v10/orm"
)

const TransactionKey = "pg_tx"

func GetConnection(ctx context.Context, db orm.DB) orm.DB {
	value := ctx.Value(TransactionKey)
	if value != nil {
		transaction, ok := value.(orm.DB)
		if ok {
			return transaction
		}
	}
	return db
}
