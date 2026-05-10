package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func SetPool(p *pgxpool.Pool) {
	pool = p
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"imageUrl"`
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS demo_products (
    id          text    PRIMARY KEY,
    name        text    NOT NULL,
    description text    NOT NULL DEFAULT '',
    price       numeric NOT NULL DEFAULT 0
);

ALTER TABLE demo_products ADD COLUMN IF NOT EXISTS image_url text NOT NULL DEFAULT '';
`

func InitSchema(ctx context.Context, p *pgxpool.Pool) error {
	_, err := p.Exec(ctx, schemaSQL)
	return err
}
