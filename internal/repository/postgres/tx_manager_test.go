package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"testing"

	"avito-backend-trainee-autumn-2025/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxManager_SuccessCommit(t *testing.T) {
	ctx := context.Background()

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	m := NewTxManager(testPool)

	err := m.WithinTx(ctx, func(ctx context.Context, repos *domain.Repos) error {

		err := repos.Team.Create(ctx, "tx_team")
		require.NoError(t, err)

		return nil
	})

	require.NoError(t, err)

	var exists bool
	err = testPool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE name = 'tx_team')`,
	).Scan(&exists)

	require.NoError(t, err)
	require.True(t, exists)
}

func TestTxManager_RollbackOnError(t *testing.T) {
	ctx := context.Background()

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	m := NewTxManager(testPool)

	err := m.WithinTx(ctx, func(ctx context.Context, repos *domain.Repos) error {
		err := repos.Team.Create(ctx, "rollback_team")
		require.NoError(t, err)

		return assert.AnError
	})

	require.Error(t, err)

	var exists bool
	err = testPool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE name = 'rollback_team')`,
	).Scan(&exists)

	require.NoError(t, err)
	require.False(t, exists)
}
