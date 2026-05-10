package handlers

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

type SeedUser struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	Email     string `yaml:"email"`
	Phone     string `yaml:"phone"`
	Address   string `yaml:"address"`
	AvatarURL string `yaml:"avatar_url"`
	Bio       string `yaml:"bio"`
	JoinedAt  string `yaml:"joined_at"`
}

type usersFile struct {
	Users []SeedUser `yaml:"users"`
}

func LoadUsers(path string) ([]SeedUser, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var data usersFile
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return data.Users, nil
}

func SeedUsers(ctx context.Context, p *pgxpool.Pool, users []SeedUser) error {
	const stmt = `
		INSERT INTO demo_users (id, name, email, phone, address, avatar_url, bio, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		    SET name       = EXCLUDED.name,
		        email      = EXCLUDED.email,
		        phone      = EXCLUDED.phone,
		        address    = EXCLUDED.address,
		        avatar_url = EXCLUDED.avatar_url,
		        bio        = EXCLUDED.bio,
		        joined_at  = EXCLUDED.joined_at`

	for _, u := range users {
		if _, err := p.Exec(ctx, stmt,
			u.ID, u.Name, u.Email, u.Phone, u.Address, u.AvatarURL, u.Bio, u.JoinedAt,
		); err != nil {
			return fmt.Errorf("upsert user %s: %w", u.ID, err)
		}
	}
	return nil
}
