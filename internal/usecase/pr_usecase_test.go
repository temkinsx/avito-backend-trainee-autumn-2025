package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestPRUsecaseCreateWithReviewers_AssignsUpToTwo(t *testing.T) {
	inserted := make(map[string]struct{})

	prRepo := &prRepositoryMock{
		createFn: func(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
			return &domain.PullRequest{
				ID:       pr.ID,
				Name:     pr.Name,
				AuthorID: pr.AuthorID,
				Status:   pr.Status,
			}, nil
		},
		insertReviewerFn: func(ctx context.Context, prID, userID string) error {
			inserted[userID] = struct{}{}
			return nil
		},
	}

	userRepo := &userRepositoryMock{
		fetchByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, TeamName: "team"}, nil
		},
		fetchActiveFn: func(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error) {
			return []*domain.User{
				{ID: "u2", TeamName: teamName, IsActive: true},
				{ID: "u3", TeamName: teamName, IsActive: true},
				{ID: "u4", TeamName: teamName, IsActive: true},
			}, nil
		},
	}

	tx := &txManagerStub{
		repos: &domain.Repos{
			PR:   prRepo,
			User: userRepo,
		},
	}

	uc := NewPRUsecase(userRepo, prRepo, tx)

	pr, err := uc.CreateWithReviewers(context.Background(), &domain.PullRequest{
		ID:       "pr1",
		Name:     "test",
		AuthorID: "author",
	})
	if err != nil {
		t.Fatalf("CreateWithReviewers: %v", err)
	}

	if len(pr.Reviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %v", pr.Reviewers)
	}

	for _, id := range pr.Reviewers {
		if _, ok := inserted[id]; !ok {
			t.Fatalf("reviewer %s was not inserted", id)
		}
	}
}

func TestPRUsecaseCreateWithReviewers_NoCandidates(t *testing.T) {
	prRepo := &prRepositoryMock{
		createFn: func(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
			return pr, nil
		},
		insertReviewerFn: func(ctx context.Context, prID, userID string) error {
			t.Fatalf("should not insert reviewer")
			return nil
		},
	}

	userRepo := &userRepositoryMock{
		fetchByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, TeamName: "team"}, nil
		},
		fetchActiveFn: func(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error) {
			return nil, nil
		},
	}

	tx := &txManagerStub{repos: &domain.Repos{PR: prRepo, User: userRepo}}

	uc := NewPRUsecase(userRepo, prRepo, tx)
	pr, err := uc.CreateWithReviewers(context.Background(), &domain.PullRequest{ID: "pr2", AuthorID: "author"})
	if err != nil {
		t.Fatalf("CreateWithReviewers: %v", err)
	}
	if len(pr.Reviewers) != 0 {
		t.Fatalf("expected no reviewers, got %v", pr.Reviewers)
	}
}

func TestPRUsecaseReassign_UsesReviewerTeam(t *testing.T) {
	var teamAsked string
	prRepo := &prRepositoryMock{
		fetchByIDFn: func(ctx context.Context, prID string) (*domain.PullRequest, error) {
			return &domain.PullRequest{ID: prID, AuthorID: "author", Status: domain.StatusOpen}, nil
		},
		reviewerAssignedFn: func(ctx context.Context, prID, userID string) (bool, error) {
			return true, nil
		},
		replaceReviewerFn: func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
			if newReviewerID != "cand1" {
				t.Fatalf("unexpected new reviewer: %s", newReviewerID)
			}
			return nil
		},
		listReviewersFn: func(ctx context.Context, prID string) ([]string, error) {
			return []string{"old", "other"}, nil
		},
	}
	userRepo := &userRepositoryMock{
		fetchByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, TeamName: "other-team"}, nil
		},
		fetchActiveFn: func(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error) {
			teamAsked = teamName
			return []*domain.User{{ID: "cand1", TeamName: teamName, IsActive: true}}, nil
		},
	}
	tx := &txManagerStub{repos: &domain.Repos{PR: prRepo, User: userRepo}}
	uc := NewPRUsecase(userRepo, prRepo, tx)

	pr, newID, err := uc.Reassign(context.Background(), "pr1", "old")
	if err != nil {
		t.Fatalf("Reassign: %v", err)
	}
	if newID != "cand1" {
		t.Fatalf("unexpected new reviewer returned: %s", newID)
	}
	if teamAsked != "other-team" {
		t.Fatalf("expected lookup in reviewer team, got %s", teamAsked)
	}
	if !reflect.DeepEqual(pr.Reviewers, []string{"old", "other"}) {
		t.Fatalf("unexpected reviewers slice: %v", pr.Reviewers)
	}
}

func TestPRUsecaseReassign_NoCandidate(t *testing.T) {
	prRepo := &prRepositoryMock{
		fetchByIDFn: func(ctx context.Context, prID string) (*domain.PullRequest, error) {
			return &domain.PullRequest{ID: prID, Status: domain.StatusOpen}, nil
		},
		reviewerAssignedFn: func(ctx context.Context, prID, userID string) (bool, error) {
			return true, nil
		},
		listReviewersFn: func(ctx context.Context, prID string) ([]string, error) {
			return []string{"old"}, nil
		},
	}
	userRepo := &userRepositoryMock{
		fetchByIDFn: func(ctx context.Context, id string) (*domain.User, error) {
			return &domain.User{ID: id, TeamName: "team"}, nil
		},
		fetchActiveFn: func(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error) {
			return nil, nil
		},
	}
	tx := &txManagerStub{repos: &domain.Repos{PR: prRepo, User: userRepo}}
	uc := NewPRUsecase(userRepo, prRepo, tx)

	_, _, err := uc.Reassign(context.Background(), "pr1", "old")
	if !errors.Is(err, domain.ErrNoCandidate) {
		t.Fatalf("expected ErrNoCandidate, got %v", err)
	}
}

func TestPRUsecaseMerge_ReturnsReviewers(t *testing.T) {
	prRepo := &prRepositoryMock{
		updateStatusFn: func(ctx context.Context, prID string) (*domain.PullRequest, error) {
			return &domain.PullRequest{ID: prID, Status: domain.StatusMerged}, nil
		},
		listReviewersFn: func(ctx context.Context, prID string) ([]string, error) {
			return []string{"u2", "u3"}, nil
		},
	}
	tx := &txManagerStub{repos: &domain.Repos{PR: prRepo}}
	uc := NewPRUsecase(nil, prRepo, tx)

	pr, err := uc.Merge(context.Background(), "pr1")
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !reflect.DeepEqual(pr.Reviewers, []string{"u2", "u3"}) {
		t.Fatalf("expected reviewers in merge response, got %v", pr.Reviewers)
	}
}
