package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestUserUsecaseSetIsActive_Success(t *testing.T) {
	userRepo := &userRepositoryMock{
		updateIsActiveFn: func(ctx context.Context, userID string, active bool) (*domain.User, error) {
			if userID != "u1" || !active {
				t.Fatalf("unexpected params: %s %v", userID, active)
			}
			return &domain.User{ID: userID, IsActive: active}, nil
		},
	}
	uc := NewUserUsecase(userRepo, nil, nil)

	user, err := uc.SetIsActive(context.Background(), "u1", true)
	if err != nil {
		t.Fatalf("SetIsActive: %v", err)
	}
	if !user.IsActive {
		t.Fatalf("expected active user")
	}
}

func TestUserUsecaseSetIsActive_NotFound(t *testing.T) {
	userRepo := &userRepositoryMock{
		updateIsActiveFn: func(ctx context.Context, userID string, active bool) (*domain.User, error) {
			return nil, domain.ErrNotFound
		},
	}
	uc := NewUserUsecase(userRepo, nil, nil)

	_, err := uc.SetIsActive(context.Background(), "missing", false)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserUsecaseGetReview_Success(t *testing.T) {
	userRepo := &userRepositoryMock{
		existsFn: func(ctx context.Context, userID string) (bool, error) {
			return true, nil
		},
	}
	prRepo := &prRepositoryMock{
		listReviewableFn: func(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
			return []*domain.PullRequest{{ID: "pr1"}, {ID: "pr2"}}, nil
		},
	}

	uc := NewUserUsecase(userRepo, prRepo, nil)

	prs, err := uc.GetReview(context.Background(), "u1")
	if err != nil {
		t.Fatalf("GetReview: %v", err)
	}
	if len(prs) != 2 || !reflect.DeepEqual(prs[0].ID, "pr1") {
		t.Fatalf("unexpected prs: %+v", prs)
	}
}

func TestUserUsecaseGetReview_NotFound(t *testing.T) {
	userRepo := &userRepositoryMock{
		existsFn: func(ctx context.Context, userID string) (bool, error) {
			return false, nil
		},
	}
	uc := NewUserUsecase(userRepo, nil, nil)

	_, err := uc.GetReview(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
