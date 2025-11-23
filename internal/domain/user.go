package domain

import "context"

type User struct {
	ID       string
	Name     string
	TeamName string
	IsActive bool
}

type UserRepository interface {
	Upsert(ctx context.Context, user *User) error
	FetchByID(ctx context.Context, id string) (*User, error)
	FetchByTeam(ctx context.Context, teamName string) ([]*User, error)
	FetchActiveByTeam(ctx context.Context, teamName string, excludeIDs ...string) ([]*User, error)
	Exists(ctx context.Context, userID string) (bool, error)
	UpdateIsActive(ctx context.Context, userID string, active bool) (*User, error)
}

type UserUsecase interface {
	SetIsActive(ctx context.Context, userID string, active bool) (*User, error)
	GetReview(ctx context.Context, userID string) ([]*PullRequest, error)
}
