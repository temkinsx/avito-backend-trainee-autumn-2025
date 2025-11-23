package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
)

type txManagerStub struct {
	repos    *domain.Repos
	withinFn func(ctx context.Context, fn func(context.Context, *domain.Repos) error) error
}

func (t *txManagerStub) WithinTx(ctx context.Context, fn func(context.Context, *domain.Repos) error) error {
	if t.withinFn != nil {
		return t.withinFn(ctx, fn)
	}
	return fn(ctx, t.repos)
}

type prRepositoryMock struct {
	createFn           func(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error)
	fetchByIDFn        func(ctx context.Context, prID string) (*domain.PullRequest, error)
	updateStatusFn     func(ctx context.Context, prID string) (*domain.PullRequest, error)
	listReviewableFn   func(ctx context.Context, userID string) ([]*domain.PullRequest, error)
	listReviewersFn    func(ctx context.Context, prID string) ([]string, error)
	insertReviewerFn   func(ctx context.Context, prID, userID string) error
	replaceReviewerFn  func(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	reviewerAssignedFn func(ctx context.Context, prID, userID string) (bool, error)
}

func (m *prRepositoryMock) Create(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
	return m.createFn(ctx, pr)
}

func (m *prRepositoryMock) FetchByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return m.fetchByIDFn(ctx, prID)
}

func (m *prRepositoryMock) UpdateStatusMerged(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return m.updateStatusFn(ctx, prID)
}

func (m *prRepositoryMock) ListReviewableByUserID(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	return m.listReviewableFn(ctx, userID)
}

func (m *prRepositoryMock) ListReviewers(ctx context.Context, prID string) ([]string, error) {
	return m.listReviewersFn(ctx, prID)
}

func (m *prRepositoryMock) InsertReviewer(ctx context.Context, prID, userID string) error {
	return m.insertReviewerFn(ctx, prID, userID)
}

func (m *prRepositoryMock) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	return m.replaceReviewerFn(ctx, prID, oldReviewerID, newReviewerID)
}

func (m *prRepositoryMock) ReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	return m.reviewerAssignedFn(ctx, prID, userID)
}

type userRepositoryMock struct {
	upsertFn         func(ctx context.Context, user *domain.User) error
	fetchByIDFn      func(ctx context.Context, id string) (*domain.User, error)
	fetchByTeamFn    func(ctx context.Context, teamName string) ([]*domain.User, error)
	fetchActiveFn    func(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error)
	existsFn         func(ctx context.Context, userID string) (bool, error)
	updateIsActiveFn func(ctx context.Context, userID string, active bool) (*domain.User, error)
}

func (m *userRepositoryMock) Upsert(ctx context.Context, user *domain.User) error {
	return m.upsertFn(ctx, user)
}

func (m *userRepositoryMock) FetchByID(ctx context.Context, id string) (*domain.User, error) {
	return m.fetchByIDFn(ctx, id)
}

func (m *userRepositoryMock) FetchByTeam(ctx context.Context, teamName string) ([]*domain.User, error) {
	return m.fetchByTeamFn(ctx, teamName)
}

func (m *userRepositoryMock) FetchActiveByTeam(ctx context.Context, teamName string, excludeIDs ...string) ([]*domain.User, error) {
	return m.fetchActiveFn(ctx, teamName, excludeIDs...)
}

func (m *userRepositoryMock) Exists(ctx context.Context, userID string) (bool, error) {
	return m.existsFn(ctx, userID)
}

func (m *userRepositoryMock) UpdateIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	return m.updateIsActiveFn(ctx, userID, active)
}

type teamRepositoryMock struct {
	createFn func(ctx context.Context, teamName string) error
	existsFn func(ctx context.Context, teamName string) (bool, error)
}

func (m *teamRepositoryMock) Create(ctx context.Context, teamName string) error {
	return m.createFn(ctx, teamName)
}

func (m *teamRepositoryMock) Exists(ctx context.Context, teamName string) (bool, error) {
	return m.existsFn(ctx, teamName)
}
