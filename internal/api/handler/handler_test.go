package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type mockPRUsecase struct {
	createFn   func(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error)
	mergeFn    func(ctx context.Context, prID string) (*domain.PullRequest, error)
	reassignFn func(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
}

func (m *mockPRUsecase) CreateWithReviewers(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
	return m.createFn(ctx, pr)
}

func (m *mockPRUsecase) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return m.mergeFn(ctx, prID)
}

func (m *mockPRUsecase) Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	return m.reassignFn(ctx, prID, oldReviewerID)
}

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

func newRecorderWithRequest(t *testing.T, method, target string, body any) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req, err := http.NewRequest(method, target, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return w, c
}

func decodeError(t *testing.T, body *bytes.Buffer) dto.ErrorResponseDTO {
	t.Helper()
	var resp dto.ErrorResponseDTO
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}

func TestPRHandlerCreate_Success(t *testing.T) {
	handler := &PRHandler{
		PRUsecase: &mockPRUsecase{
			createFn: func(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
				return &domain.PullRequest{
					ID:       pr.ID,
					Name:     pr.Name,
					AuthorID: pr.AuthorID,
					Status:   domain.StatusOpen,
					Reviewers: []string{
						"u2", "u3",
					},
				}, nil
			},
		},
	}

	reqBody := dto.PullRequestCreateRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "test",
		AuthorID:        "u1",
	}
	w, c := newRecorderWithRequest(t, http.MethodPost, "/pullRequest/create", reqBody)

	handler.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", w.Code)
	}

	var resp dto.PullRequestCreateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.PR.PullRequestID != reqBody.PullRequestID || len(resp.PR.AssignedReviewers) != 2 {
		t.Fatalf("unexpected response: %+v", resp.PR)
	}
}

func TestPRHandlerCreate_NotFound(t *testing.T) {
	handler := &PRHandler{
		PRUsecase: &mockPRUsecase{
			createFn: func(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
				return nil, domain.ErrNotFound
			},
		},
	}

	w, c := newRecorderWithRequest(t, http.MethodPost, "/pullRequest/create", dto.PullRequestCreateRequest{
		PullRequestID:   "pr-unknown",
		PullRequestName: "test",
		AuthorID:        "missing",
	})

	handler.Create(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}

	resp := decodeError(t, w.Body)
	if resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("unexpected error code: %+v", resp)
	}
}

func TestPRHandlerMerge_NotFound(t *testing.T) {
	handler := &PRHandler{
		PRUsecase: &mockPRUsecase{
			mergeFn: func(ctx context.Context, prID string) (*domain.PullRequest, error) {
				return nil, domain.ErrNotFound
			},
		},
	}

	w, c := newRecorderWithRequest(t, http.MethodPost, "/pullRequest/merge", dto.PullRequestMergeRequest{PullRequestID: "missing"})

	handler.Merge(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPRHandlerReassign_NoCandidate(t *testing.T) {
	handler := &PRHandler{
		PRUsecase: &mockPRUsecase{
			reassignFn: func(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
				return nil, "", domain.ErrNoCandidate
			},
		},
	}

	w, c := newRecorderWithRequest(t, http.MethodPost, "/pullRequest/reassign", dto.PullRequestReassignRequest{
		PullRequestID: "pr1",
		OldUserID:     "u2",
	})

	handler.Reassign(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}

	resp := decodeError(t, w.Body)
	if resp.Error.Code != "NO_CANDIDATE" {
		t.Fatalf("unexpected error code: %+v", resp)
	}
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
