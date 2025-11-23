package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
)

type teamRepository struct {
	q Querier
}

func NewTeamRepository(q Querier) domain.TeamRepository {
	return &teamRepository{q: q}
}

func (tr *teamRepository) Create(ctx context.Context, teamName string) error {
	const q = `
		INSERT INTO teams (name)
		VALUES ($1)
	`

	_, err := tr.q.Exec(ctx, q, teamName)
	if err != nil {
		return err
	}

	return nil
}

func (tr *teamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	const q = `
        SELECT EXISTS (
            SELECT 1
            FROM teams
            WHERE name = $1
        );
    `

	var exists bool
	err := tr.q.QueryRow(ctx, q, teamName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
