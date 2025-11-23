package domain

import (
	"context"
	"time"
)

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PRStatus
	Reviewers []string
	CreatedAt *time.Time
	MergedAt  *time.Time
}

type PRRepository interface {
	Create(ctx context.Context, pr *PullRequest) (*PullRequest, error)
	FetchByID(ctx context.Context, prID string) (*PullRequest, error)
	UpdateStatusMerged(ctx context.Context, prID string) (*PullRequest, error)
	ListReviewableByUserID(ctx context.Context, userID string) ([]*PullRequest, error)
	ListReviewers(ctx context.Context, prID string) ([]string, error)
	InsertReviewer(ctx context.Context, prID, userID string) error
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	ReviewerAssigned(ctx context.Context, prID, userID string) (bool, error)
}

type PRUsecase interface {
	CreateWithReviewers(ctx context.Context, newPR *PullRequest) (*PullRequest, error)
	Merge(ctx context.Context, prID string) (*PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*PullRequest, string, error)
}
