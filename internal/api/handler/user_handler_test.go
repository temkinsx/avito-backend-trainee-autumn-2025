package handler

import (
	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"net/http"
	"testing"
)

type mockUserUsecase struct {
	setActiveFn func(ctx context.Context, userID string, active bool) (*domain.User, error)
	getReviewFn func(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}

func (m *mockUserUsecase) SetIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	return m.setActiveFn(ctx, userID, active)
}

func (m *mockUserUsecase) GetReview(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	return m.getReviewFn(ctx, userID)
}

func TestUserHandlerSetIsActive_NotFound(t *testing.T) {
	handler := &UserHandler{
		UserUsecase: &mockUserUsecase{
			setActiveFn: func(ctx context.Context, userID string, active bool) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
		},
	}

	w, c := newRecorderWithRequest(t, http.MethodPost, "/users/setIsActive", dto.UsersSetIsActiveRequest{
		UserID:   "missing",
		IsActive: false,
	})

	handler.SetIsActive(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUserHandlerGetReview_Validation(t *testing.T) {
	handler := &UserHandler{
		UserUsecase: &mockUserUsecase{},
	}

	w, c := newRecorderWithRequest(t, http.MethodGet, "/users/getReview", nil)

	handler.GetReview(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
