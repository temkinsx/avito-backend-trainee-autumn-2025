package handler

import (
	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

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
