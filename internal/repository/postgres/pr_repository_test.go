package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"avito-backend-trainee-autumn-2025/testutils"
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestPRRepository_Create_Success(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	newPR := &domain.PullRequest{
		ID:       "pr_new",
		Name:     "new feature",
		AuthorID: testutils.User1ID,
		Status:   domain.StatusOpen,
	}

	created, err := repo.Create(ctx, newPR)
	require.NoError(t, err)
	require.NotNil(t, created)

	require.Equal(t, newPR.ID, created.ID)
	require.Equal(t, newPR.Name, created.Name)
	require.Equal(t, newPR.AuthorID, created.AuthorID)
	require.Equal(t, newPR.Status, created.Status)
	require.NotNil(t, created.CreatedAt)

	var dbPR domain.PullRequest
	err = testPool.QueryRow(ctx, `
		SELECT id, name, author_id, status
		FROM pull_requests
		WHERE id = $1
	`, newPR.ID).Scan(
		&dbPR.ID,
		&dbPR.Name,
		&dbPR.AuthorID,
		&dbPR.Status,
	)
	require.NoError(t, err)
	require.Equal(t, newPR.ID, dbPR.ID)
	require.Equal(t, newPR.Name, dbPR.Name)
	require.Equal(t, newPR.AuthorID, dbPR.AuthorID)
	require.Equal(t, newPR.Status, dbPR.Status)
}

func TestPRRepository_Create_ErrAlreadyExists(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	pr := &domain.PullRequest{
		ID:       testutils.PR1ID,
		Name:     "whatever",
		AuthorID: testutils.User1ID,
		Status:   domain.StatusOpen,
	}

	created, err := repo.Create(ctx, pr)
	require.Error(t, err)
	require.Nil(t, created)
	require.ErrorIs(t, err, domain.ErrAlreadyExists)

	var name string
	err = testPool.QueryRow(ctx, `
		SELECT name FROM pull_requests WHERE id = $1
	`, testutils.PR1ID).Scan(&name)
	require.NoError(t, err)
	require.Equal(t, testutils.PR1Name, name)
}

func TestPRRepository_Create_ErrNotFoundAuthor(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	const newID = "pr_bad_author"

	pr := &domain.PullRequest{
		ID:       newID,
		Name:     "bad author",
		AuthorID: "no_such_user",
		Status:   domain.StatusOpen,
	}

	created, err := repo.Create(ctx, pr)
	require.Error(t, err)
	require.Nil(t, created)
	require.ErrorIs(t, err, domain.ErrNotFound)

	var count int
	err = testPool.QueryRow(ctx, `
		SELECT COUNT(*) FROM pull_requests WHERE id = $1
	`, newID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestPRRepository_FetchByID(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("success", func(t *testing.T) {
		pr, err := repo.FetchByID(ctx, testutils.PR1ID)
		require.NoError(t, err)
		require.NotNil(t, pr)

		require.Equal(t, testutils.PR1ID, pr.ID)
		require.Equal(t, testutils.PR1Name, pr.Name)
		require.Equal(t, testutils.User1ID, pr.AuthorID)
		require.Equal(t, domain.StatusOpen, pr.Status)
		require.NotNil(t, pr.CreatedAt)
		require.Nil(t, pr.MergedAt)
	})

	t.Run("not_found", func(t *testing.T) {
		pr, err := repo.FetchByID(ctx, "no_such_pr")
		require.Error(t, err)
		require.Nil(t, pr)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestPRRepository_UpdateStatusMerged(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("merge_open_pr", func(t *testing.T) {
		pr, err := repo.UpdateStatusMerged(ctx, testutils.PR1ID)
		require.NoError(t, err)
		require.NotNil(t, pr)

		require.Equal(t, testutils.PR1ID, pr.ID)
		require.Equal(t, domain.StatusMerged, pr.Status)
		require.NotNil(t, pr.MergedAt)

		var status string
		err = testPool.QueryRow(ctx, `
			SELECT status FROM pull_requests WHERE id = $1
		`, testutils.PR1ID).Scan(&status)
		require.NoError(t, err)
		require.Equal(t, string(domain.StatusMerged), status)
	})

	t.Run("not_found", func(t *testing.T) {
		pr, err := repo.UpdateStatusMerged(ctx, "no_such_pr")
		require.Error(t, err)
		require.Nil(t, pr)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestPRRepository_ListReviewableByUserID(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("user2_has_two_prs", func(t *testing.T) {
		prs, err := repo.ListReviewableByUserID(ctx, testutils.User2ID)
		require.NoError(t, err)
		require.Len(t, prs, 2)

		ids := []string{prs[0].ID, prs[1].ID}
		require.ElementsMatch(t, []string{testutils.PR1ID, testutils.PR2ID}, ids)
	})

	t.Run("user5_has_one_pr", func(t *testing.T) {
		prs, err := repo.ListReviewableByUserID(ctx, testutils.User5ID)
		require.NoError(t, err)
		require.Len(t, prs, 1)
		require.Equal(t, testutils.PR3ID, prs[0].ID)
	})

	t.Run("user1_has_no_prs", func(t *testing.T) {
		prs, err := repo.ListReviewableByUserID(ctx, testutils.User1ID)
		require.NoError(t, err)
		require.Len(t, prs, 0)
	})
}

func TestPRRepository_ListReviewers(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("pr1_reviewers", func(t *testing.T) {
		revs, err := repo.ListReviewers(ctx, testutils.PR1ID)
		require.NoError(t, err)
		require.ElementsMatch(t, []string{testutils.User2ID, testutils.User3ID}, revs)
	})

	t.Run("no_reviewers_for_unknown_pr", func(t *testing.T) {
		revs, err := repo.ListReviewers(ctx, "no_such_pr")
		require.NoError(t, err)
		require.Len(t, revs, 0)
	})
}

func TestPRRepository_InsertReviewer(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	err := repo.InsertReviewer(ctx, testutils.PR3ID, testutils.User2ID)
	require.NoError(t, err)

	rows, err := testPool.Query(ctx, `
		SELECT user_id
		FROM pr_reviewers
		WHERE pr_id = $1
	`, testutils.PR3ID)
	require.NoError(t, err)
	defer rows.Close()

	revs, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (string, error) {
		var id string
		if err := r.Scan(&id); err != nil {
			return "", err
		}
		return id, nil
	})
	require.NoError(t, err)

	require.ElementsMatch(t, []string{testutils.User5ID, testutils.User2ID}, revs)
}

func TestPRRepository_ReplaceReviewer(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	_, err := testPool.Exec(ctx, `
		INSERT INTO users (id, name, team_name, is_active)
		VALUES ($1, 'Extra User', $2, TRUE)
	`, "u_extra", testutils.TestTeam)
	require.NoError(t, err)

	err = repo.ReplaceReviewer(ctx, testutils.PR1ID, testutils.User2ID, "u_extra")
	require.NoError(t, err)

	rows, err := testPool.Query(ctx, `
		SELECT user_id
		FROM pr_reviewers
		WHERE pr_id = $1
	`, testutils.PR1ID)
	require.NoError(t, err)
	defer rows.Close()

	revs, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (string, error) {
		var id string
		if err := r.Scan(&id); err != nil {
			return "", err
		}
		return id, nil
	})
	require.NoError(t, err)

	require.ElementsMatch(t, []string{"u_extra", testutils.User3ID}, revs)
	require.NotContains(t, revs, testutils.User2ID)
}

func TestPRRepository_ReviewerAssigned(t *testing.T) {
	ctx := context.Background()
	repo := NewPRRepository(testPool)

	require.NoError(t, testutils.PrepareTestTablesWithFixtures(ctx, testPool))

	t.Run("assigned_true", func(t *testing.T) {
		ok, err := repo.ReviewerAssigned(ctx, testutils.PR1ID, testutils.User2ID)
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("assigned_false", func(t *testing.T) {
		ok, err := repo.ReviewerAssigned(ctx, testutils.PR1ID, testutils.User4ID)
		require.NoError(t, err)
		require.False(t, ok)
	})
}
