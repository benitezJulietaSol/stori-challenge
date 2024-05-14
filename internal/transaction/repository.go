package transaction

import (
	"awesomeProject2/internal/integrations/db"
	"awesomeProject2/internal/model"
	"context"
	"github.com/go-pg/pg/v10/orm"
)

type Repository struct {
	db orm.DB
}

func NewRepository(db orm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertTransactions(ctx context.Context, transactions []model.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	database := db.GetConnection(ctx, r.db)
	if _, err := database.ModelContext(ctx, &transactions).Insert(); err != nil {
		return err
	}

	return nil
}
