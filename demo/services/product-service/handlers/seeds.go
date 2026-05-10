package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

type SeedProduct struct {
	ID          string  `yaml:"id"`
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Price       float64 `yaml:"price"`
	ImageURL    string  `yaml:"image_url"`
	Category    string  `yaml:"category"`
}

type productsFile struct {
	Products []SeedProduct `yaml:"products"`
}

func LoadProducts(path string) ([]SeedProduct, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var data productsFile
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return data.Products, nil
}

func SeedProducts(ctx context.Context, p *pgxpool.Pool, products []SeedProduct) error {
	const stmt = `
		INSERT INTO demo_products (id, name, description, price, image_url)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		    SET name        = EXCLUDED.name,
		        description = EXCLUDED.description,
		        price       = EXCLUDED.price,
		        image_url   = EXCLUDED.image_url`

	for _, sp := range products {
		if _, err := p.Exec(ctx, stmt, sp.ID, sp.Name, sp.Description, sp.Price, sp.ImageURL); err != nil {
			return fmt.Errorf("upsert product %s: %w", sp.ID, err)
		}
	}
	return nil
}
