package handlers

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type seedProduct struct {
	ID       string `yaml:"id"`
	Category string `yaml:"category"`
}

type productsFile struct {
	Products []seedProduct `yaml:"products"`
}

func LoadCatalog(path string) ([]CatalogEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var data productsFile
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make([]CatalogEntry, 0, len(data.Products))
	for _, p := range data.Products {
		out = append(out, CatalogEntry{ID: p.ID, Category: p.Category})
	}
	return out, nil
}
