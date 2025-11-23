package domain

import "context"

type Team struct {
	Name    string
	Members []*User
}

type TeamRepository interface {
	Create(ctx context.Context, teamName string) error
	Exists(ctx context.Context, teamName string) (bool, error)
}

type TeamUsecase interface {
	Add(ctx context.Context, team *Team) (*Team, error)
	ListByName(ctx context.Context, name string) (*Team, error)
}
