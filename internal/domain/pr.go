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
	// Create inserts new PR (without reviewers)
	Create(ctx context.Context, pr *PullRequest) (*PullRequest, error)
	// FetchByID returns PR without reviewers
	FetchByID(ctx context.Context, prID string) (*PullRequest, error)
	// UpdateStatusMerged sets PR.status = MERGED and returns updated PR
	UpdateStatusMerged(ctx context.Context, prID string) (*PullRequest, error)
	// ListReviewableByUserID returns PRs where userID is assigned reviewer
	ListReviewableByUserID(ctx context.Context, userID string) ([]*PullRequest, error)
	// ListReviewers returns all reviewers for PR
	ListReviewers(ctx context.Context, prID string) ([]string, error)
	// InsertReviewer adds user to PR reviewers
	InsertReviewer(ctx context.Context, prID, userID string) error
	// ReplaceReviewer old → new
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	// ReviewerAssigned checks if user is reviewer on PR
	ReviewerAssigned(ctx context.Context, prID, userID string) (bool, error)
}

type PRUsecase interface {
	CreateWithReviewers(ctx context.Context, newPR *PullRequest) (*PullRequest, error)
	Merge(ctx context.Context, prID string) (*PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*PullRequest, string, error)
}
