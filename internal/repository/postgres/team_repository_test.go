package postgres

import (
	"avito-backend-trainee-autumn-2025/testutils"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTeamRepository_Create(t *testing.T) {
	ctx := context.Background()
	tr := NewTeamRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("success_create_new_team", func(t *testing.T) {
		teamName := "new_team"

		err := tr.Create(ctx, teamName)
		require.NoError(t, err)

		var exists bool
		err = testPool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`,
			teamName,
		).Scan(&exists)

		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("error_team_already_exists", func(t *testing.T) {
		err := tr.Create(ctx, testutils.TestTeam)
		require.Error(t, err)

		require.Contains(t, err.Error(), "duplicate")
	})
}

func TestTeamRepository_Exists(t *testing.T) {
	ctx := context.Background()
	tr := NewTeamRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("team_exists", func(t *testing.T) {
		ok, err := tr.Exists(ctx, testutils.TestTeam)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("team_not_exists", func(t *testing.T) {
		ok, err := tr.Exists(ctx, "unknown_team")
		require.NoError(t, err)
		require.False(t, ok)
	})
}
