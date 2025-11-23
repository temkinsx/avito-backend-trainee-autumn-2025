package handler

import (
	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"net/http"
	"testing"
)

type mockTeamUsecase struct {
	addFn       func(ctx context.Context, team *domain.Team) (*domain.Team, error)
	listByNameF func(ctx context.Context, name string) (*domain.Team, error)
}

func (m *mockTeamUsecase) Add(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	return m.addFn(ctx, team)
}

func (m *mockTeamUsecase) ListByName(ctx context.Context, name string) (*domain.Team, error) {
	return m.listByNameF(ctx, name)
}

func TestTeamHandlerAdd_Duplicate(t *testing.T) {
	handler := &TeamHandler{
		TeamUsecase: &mockTeamUsecase{
			addFn: func(ctx context.Context, team *domain.Team) (*domain.Team, error) {
				return nil, domain.ErrAlreadyExists
			},
		},
	}

	w, c := newRecorderWithRequest(t, http.MethodPost, "/team/add", dto.TeamAddRequest{
		TeamName: "team",
		Members: []dto.TeamMemberDTO{
			{UserID: "u1", Username: "One", IsActive: true},
		},
	})

	handler.Add(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	resp := decodeError(t, w.Body)
	if resp.Error.Code != "TEAM_EXISTS" {
		t.Fatalf("unexpected error code: %+v", resp)
	}
}

func TestTeamHandlerGet_Validation(t *testing.T) {
	handler := &TeamHandler{
		TeamUsecase: &mockTeamUsecase{},
	}

	w, c := newRecorderWithRequest(t, http.MethodGet, "/team/get", nil)

	handler.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
