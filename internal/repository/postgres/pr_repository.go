package postgres

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type prRepository struct {
	q Querier
}

func NewPRRepository(q Querier) domain.PRRepository {
	return &prRepository{q: q}
}

func (p *prRepository) Create(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
	const q = `
        INSERT INTO pull_requests (id, name, author_id, status)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, author_id, status, created_at;
    `

	var created domain.PullRequest
	err := p.q.QueryRow(ctx, q, pr.ID, pr.Name, pr.AuthorID, pr.Status).Scan(
		&created.ID, &created.Name,
		&created.AuthorID, &created.Status,
		&created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, domain.ErrAlreadyExists
			case "23503":
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}

	return &created, nil
}

func (p *prRepository) FetchByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	const q = `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1;
	`

	var pr domain.PullRequest
	err := p.q.QueryRow(ctx, q, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &pr, nil
}

func (p *prRepository) UpdateStatusMerged(ctx context.Context, prID string) (*domain.PullRequest, error) {
	const q = `
		UPDATE pull_requests
			SET status = 'MERGED',
			    merged_at = COALESCE(merged_at, now())
		WHERE id = $1
		RETURNING id, name, author_id, status, created_at, merged_at;
	`

	var pr domain.PullRequest
	err := p.q.QueryRow(ctx, q, prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &pr, nil
}

func (p *prRepository) ListReviewableByUserID(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	const q = `
		SELECT p.id, p.name, p.author_id, p.status, p.created_at, p.merged_at
		FROM pr_reviewers r
		JOIN pull_requests p ON p.id = r.pr_id
		WHERE r.user_id = $1;
	`

	rows, err := p.q.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prs, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (*domain.PullRequest, error) {
		var pr domain.PullRequest
		if err := r.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		return &pr, nil
	})
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (p *prRepository) ListReviewers(ctx context.Context, prID string) ([]string, error) {
	const q = `
		SELECT user_id
		FROM pr_reviewers
		WHERE pr_id = $1;
	`

	rows, err := p.q.Query(ctx, q, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revIDs, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (string, error) {
		var id string
		if err := r.Scan(&id); err != nil {
			return "", err
		}
		return id, nil
	})
	if err != nil {
		return nil, err
	}

	return revIDs, nil
}

func (p *prRepository) InsertReviewer(ctx context.Context, prID, userID string) error {
	const q = `
		INSERT INTO pr_reviewers (pr_id, user_id)
		VALUES ($1, $2);
	`

	_, err := p.q.Exec(ctx, q, prID, userID)
	return err
}

func (p *prRepository) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	const q = `
        UPDATE pr_reviewers
        SET user_id = $1
        WHERE pr_id = $2 AND user_id = $3
    `

	_, err := p.q.Exec(ctx, q, newReviewerID, prID, oldReviewerID)
	return err
}

func (p *prRepository) ReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM pr_reviewers
			WHERE pr_id = $1
				AND user_id = $2
		);
	`

	var exists bool
	err := p.q.QueryRow(ctx, q, prID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
