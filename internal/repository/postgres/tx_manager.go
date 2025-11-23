package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type txManager struct {
	pool *pgxpool.Pool
}

func (m *txManager) WithinTx(ctx context.Context, fn func(ctx context.Context, repos *domain.Repos) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	repos := &domain.Repos{
		PR:   NewPRRepository(tx),
		User: NewUserRepository(tx),
		Team: NewTeamRepository(tx),
	}

	if err := fn(ctx, repos); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
