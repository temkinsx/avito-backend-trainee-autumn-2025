package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"errors"
	"testing"
)

func TestTeamUsecaseAdd_Success(t *testing.T) {
	teamRepo := &teamRepositoryMock{
		existsFn: func(ctx context.Context, teamName string) (bool, error) {
			return false, nil
		},
		createFn: func(ctx context.Context, teamName string) error {
			if teamName != "team" {
				t.Fatalf("unexpected team name: %s", teamName)
			}
			return nil
		},
	}

	upserted := make([]*domain.User, 0)
	userRepo := &userRepositoryMock{
		fetchByTeamFn: func(ctx context.Context, teamName string) ([]*domain.User, error) {
			return nil, nil
		},
		upsertFn: func(ctx context.Context, user *domain.User) error {
			upserted = append(upserted, user)
			return nil
		},
	}

	tx := &txManagerStub{
		repos: &domain.Repos{
			Team: teamRepo,
			User: userRepo,
		},
	}

	uc := NewTeamUsecase(teamRepo, userRepo, tx)

	team := &domain.Team{
		Name: "team",
		Members: []*domain.User{
			{ID: "u1", Name: "One", IsActive: true},
			{ID: "u2", Name: "Two", IsActive: false},
		},
	}

	created, err := uc.Add(context.Background(), team)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if created != team {
		t.Fatalf("expected same team pointer")
	}

	if len(upserted) != 2 {
		t.Fatalf("expected 2 upserts, got %d", len(upserted))
	}
	for _, u := range upserted {
		if u.TeamName != "team" {
			t.Fatalf("upsert called with wrong team: %+v", u)
		}
	}
}

func TestTeamUsecaseAdd_Duplicate(t *testing.T) {
	teamRepo := &teamRepositoryMock{
		existsFn: func(ctx context.Context, teamName string) (bool, error) {
			return true, nil
		},
	}
	uc := NewTeamUsecase(teamRepo, nil, nil)

	_, err := uc.Add(context.Background(), &domain.Team{Name: "team"})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestTeamUsecaseListByName_Success(t *testing.T) {
	teamRepo := &teamRepositoryMock{
		existsFn: func(ctx context.Context, teamName string) (bool, error) {
			return true, nil
		},
	}
	userRepo := &userRepositoryMock{
		fetchByTeamFn: func(ctx context.Context, teamName string) ([]*domain.User, error) {
			return []*domain.User{{ID: "u1"}, {ID: "u2"}}, nil
		},
	}

	uc := NewTeamUsecase(teamRepo, userRepo, nil)

	team, err := uc.ListByName(context.Background(), "team")
	if err != nil {
		t.Fatalf("ListByName: %v", err)
	}
	if team.Name != "team" || len(team.Members) != 2 {
		t.Fatalf("unexpected team: %+v", team)
	}
}

func TestTeamUsecaseListByName_NotFound(t *testing.T) {
	teamRepo := &teamRepositoryMock{
		existsFn: func(ctx context.Context, teamName string) (bool, error) {
			return false, nil
		},
	}
	uc := NewTeamUsecase(teamRepo, nil, nil)

	_, err := uc.ListByName(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
