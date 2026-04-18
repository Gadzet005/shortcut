package failurepostgres

import (
	"context"
	"github.com/jmoiron/sqlx"
)

func NewPostgresRepo(db *sqlx.DB) (*postgresRepo, error) {
	return &postgresRepo{db: db}, nil
}

type postgresRepo struct {
	db *sqlx.DB
}

func (pr *postgresRepo) AddFail(ctx context.Context,) {

}	

func (pr *postgresRepo) GetFail(ctx context.Context,) {

}	

func (pr *postgresRepo) UpdateFail(ctx context.Context,) {

}

func (pr *postgresRepo) DeleteFail(ctx context.Context,) {

}
