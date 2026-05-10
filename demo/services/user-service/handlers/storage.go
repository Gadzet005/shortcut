package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func SetPool(p *pgxpool.Pool) {
	pool = p
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS demo_users (
    id    text PRIMARY KEY,
    name  text NOT NULL,
    email text NOT NULL DEFAULT ''
);

ALTER TABLE demo_users ADD COLUMN IF NOT EXISTS phone      text NOT NULL DEFAULT '';
ALTER TABLE demo_users ADD COLUMN IF NOT EXISTS address    text NOT NULL DEFAULT '';
ALTER TABLE demo_users ADD COLUMN IF NOT EXISTS avatar_url text NOT NULL DEFAULT '';
ALTER TABLE demo_users ADD COLUMN IF NOT EXISTS bio        text NOT NULL DEFAULT '';
ALTER TABLE demo_users ADD COLUMN IF NOT EXISTS joined_at  text NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS demo_cart_items (
    user_id    text    NOT NULL,
    product_id text    NOT NULL,
    qty        integer NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, product_id)
);
`

func InitSchema(ctx context.Context, p *pgxpool.Pool) error {
	_, err := p.Exec(ctx, schemaSQL)
	return err
}
