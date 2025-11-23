package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"avito-backend-trainee-autumn-2025/testutils"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepository_Upsert_InsertNew(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
	require.NoError(t, err)

	u := &domain.User{
		ID:       "u_new",
		Name:     "New User",
		TeamName: testutils.TestTeam,
		IsActive: true,
	}

	err = repo.Upsert(ctx, u)
	require.NoError(t, err)

	var dbUser domain.User
	err = testPool.QueryRow(ctx, `
		SELECT id, name, team_name, is_active
		FROM users
		WHERE id = $1
	`, u.ID).Scan(&dbUser.ID, &dbUser.Name, &dbUser.TeamName, &dbUser.IsActive)
	require.NoError(t, err)
	require.Equal(t, u, &dbUser)
}

func TestUserRepository_Upsert_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
	require.NoError(t, err)

	var beforeIsActive bool
	err = testPool.QueryRow(ctx, `
		SELECT is_active
		FROM users
		WHERE id = $1
	`, testutils.User1ID).Scan(&beforeIsActive)
	require.NoError(t, err)

	u := &domain.User{
		ID:       testutils.User1ID,
		Name:     "Updated Name",
		TeamName: testutils.TestTeam,
		IsActive: !beforeIsActive,
	}

	err = repo.Upsert(ctx, u)
	require.NoError(t, err)

	var dbUser domain.User
	err = testPool.QueryRow(ctx, `
		SELECT id, name, team_name, is_active
		FROM users
		WHERE id = $1
	`, u.ID).Scan(&dbUser.ID, &dbUser.Name, &dbUser.TeamName, &dbUser.IsActive)
	require.NoError(t, err)
	require.Equal(t, u, &dbUser)
}

func TestUserRepository_FetchByID(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		got, err := repo.FetchByID(ctx, testutils.User1ID)
		require.NoError(t, err)

		require.Equal(t, &domain.User{
			ID:       testutils.User1ID,
			Name:     "User One",
			TeamName: testutils.TestTeam,
			IsActive: true,
		}, got)
	})

	t.Run("not_found", func(t *testing.T) {
		got, err := repo.FetchByID(ctx, "no_such_user")
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, got)
	})
}

func TestUserRepository_FetchByTeam(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
	require.NoError(t, err)

	t.Run("test_team", func(t *testing.T) {
		users, err := repo.FetchByTeam(ctx, testutils.TestTeam)
		require.NoError(t, err)

		require.Len(t, users, 3)
		ids := []string{users[0].ID, users[1].ID, users[2].ID}
		require.ElementsMatch(t, []string{
			testutils.User1ID,
			testutils.User2ID,
			testutils.User3ID,
		}, ids)
	})

	t.Run("other_team", func(t *testing.T) {
		users, err := repo.FetchByTeam(ctx, testutils.OtherTeam)
		require.NoError(t, err)

		require.Len(t, users, 3)
		ids := []string{users[0].ID, users[1].ID, users[2].ID}
		require.ElementsMatch(t, []string{
			testutils.User4ID,
			testutils.User5ID,
			testutils.User6ID,
		}, ids)
	})
}

func TestUserRepository_FetchActiveByTeam(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
	require.NoError(t, err)

	t.Run("no_exclude", func(t *testing.T) {
		users, err := repo.FetchActiveByTeam(ctx, testutils.OtherTeam)
		require.NoError(t, err)

		require.Len(t, users, 2)
		ids := []string{users[0].ID, users[1].ID}
		require.ElementsMatch(t, []string{testutils.User4ID, testutils.User5ID}, ids)
	})

	t.Run("exclude_one", func(t *testing.T) {
		users, err := repo.FetchActiveByTeam(ctx, testutils.OtherTeam, testutils.User5ID)
		require.NoError(t, err)

		require.Len(t, users, 1)
		require.Equal(t, testutils.User4ID, users[0].ID)
	})

	t.Run("exclude_all", func(t *testing.T) {
		users, err := repo.FetchActiveByTeam(ctx, testutils.OtherTeam, testutils.User4ID, testutils.User5ID)
		require.NoError(t, err)
		require.Len(t, users, 0)
	})
}

func TestUserRepository_Exists(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
	require.NoError(t, err)

	t.Run("exists_true", func(t *testing.T) {
		ok, err := repo.Exists(ctx, testutils.User1ID)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("exists_false", func(t *testing.T) {
		ok, err := repo.Exists(ctx, "no_such_user")
		require.NoError(t, err)
		require.False(t, ok)
	})
}

func TestUserRepository_UpdateIsActive(t *testing.T) {
	ctx := context.Background()
	repo := NewUserRepository(testPool)

	t.Run("success_change_status", func(t *testing.T) {
		err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
		require.NoError(t, err)

		updated, err := repo.UpdateIsActive(ctx, testutils.User1ID, false)
		require.NoError(t, err)
		require.Equal(t, &domain.User{
			ID:       testutils.User1ID,
			Name:     "User One",
			TeamName: testutils.TestTeam,
			IsActive: false,
		}, updated)

		var dbIsActive bool
		err = testPool.QueryRow(ctx, `
			SELECT is_active
			FROM users
			WHERE id = $1
		`, testutils.User1ID).Scan(&dbIsActive)
		require.NoError(t, err)
		require.False(t, dbIsActive)
	})

	t.Run("success_same_status", func(t *testing.T) {
		err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
		require.NoError(t, err)

		updated, err := repo.UpdateIsActive(ctx, testutils.User1ID, true)
		require.NoError(t, err)
		require.Equal(t, &domain.User{
			ID:       testutils.User1ID,
			Name:     "User One",
			TeamName: testutils.TestTeam,
			IsActive: true,
		}, updated)

		var dbIsActive bool
		err = testPool.QueryRow(ctx, `
			SELECT is_active
			FROM users
			WHERE id = $1
		`, testutils.User1ID).Scan(&dbIsActive)
		require.NoError(t, err)
		require.True(t, dbIsActive)
	})

	t.Run("not_found", func(t *testing.T) {
		err := testutils.PrepareTestTablesWithFixtures(ctx, testPool)
		require.NoError(t, err)

		updated, err := repo.UpdateIsActive(ctx, "no_such_user", false)
		require.Error(t, err)
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, updated)

		var count int
		err = testPool.QueryRow(ctx, `
			SELECT COUNT(*)
			FROM users
			WHERE id = $1
		`, "no_such_user").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
}
