package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type prUsecase struct {
	userRepository domain.UserRepository
	prRepository   domain.PRRepository
	txManager      domain.TxManager
}

func NewPRUsecase(userRepository domain.UserRepository, prRepository domain.PRRepository, txManager domain.TxManager) domain.PRUsecase {
	return &prUsecase{
		userRepository: userRepository,
		prRepository:   prRepository,
		txManager:      txManager,
	}
}

func (p *prUsecase) CreateWithReviewers(ctx context.Context, newPR *domain.PullRequest) (*domain.PullRequest, error) {
	var result *domain.PullRequest

	err := p.txManager.WithinTx(ctx, func(ctx context.Context, repos *domain.Repos) error {
		if newPR.Status == "" {
			newPR.Status = domain.StatusOpen
		}

		createdPR, err := repos.PR.Create(ctx, newPR)
		if err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				return fmt.Errorf("PR id %w", err)
			}
			return err
		}

		author, err := repos.User.FetchByID(ctx, createdPR.AuthorID)
		if err != nil {
			return err
		}

		candidates, err := repos.User.FetchActiveByTeam(ctx, author.TeamName, author.ID)
		if err != nil {
			return err
		}

		if len(candidates) == 0 {
			createdPR.Reviewers = nil
			result = createdPR
			return nil
		}

		rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		if len(candidates) > 2 {
			candidates = candidates[:2]
		}

		reviewerIDs := make([]string, 0, len(candidates))
		for _, r := range candidates {
			if err := repos.PR.InsertReviewer(ctx, createdPR.ID, r.ID); err != nil {
				return err
			}
			reviewerIDs = append(reviewerIDs, r.ID)
		}

		createdPR.Reviewers = reviewerIDs
		result = createdPR
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *prUsecase) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var result *domain.PullRequest

	err := p.txManager.WithinTx(ctx, func(ctx context.Context, repos *domain.Repos) error {
		pr, err := repos.PR.UpdateStatusMerged(ctx, prID)
		if err != nil {
			return err
		}

		reviewers, err := repos.PR.ListReviewers(ctx, pr.ID)
		if err != nil {
			return err
		}

		pr.Reviewers = reviewers
		result = pr
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *prUsecase) Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	var (
		result   *domain.PullRequest
		newRevID string
	)

	err := p.txManager.WithinTx(ctx, func(ctx context.Context, repos *domain.Repos) error {
		pr, err := repos.PR.FetchByID(ctx, prID)
		if err != nil {
			return err
		}

		if pr.Status == domain.StatusMerged {
			return domain.ErrPRMerged
		}

		assigned, err := repos.PR.ReviewerAssigned(ctx, prID, oldReviewerID)
		if err != nil {
			return err
		}
		if !assigned {
			return domain.ErrNotAssigned
		}

		oldReviewer, err := repos.User.FetchByID(ctx, oldReviewerID)
		if err != nil {
			return err
		}

		currentReviewers, err := repos.PR.ListReviewers(ctx, pr.ID)
		if err != nil {
			return err
		}

		excludeIDs := make([]string, 0, len(currentReviewers)+2)
		excludeIDs = append(excludeIDs, oldReviewerID, pr.AuthorID)
		for _, reviewerID := range currentReviewers {
			if reviewerID != oldReviewerID {
				excludeIDs = append(excludeIDs, reviewerID)
			}
		}

		revs, err := repos.User.FetchActiveByTeam(ctx, oldReviewer.TeamName, excludeIDs...)
		if err != nil {
			return err
		}
		if len(revs) == 0 {
			return domain.ErrNoCandidate
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		newRevID = revs[r.Intn(len(revs))].ID

		err = repos.PR.ReplaceReviewer(ctx, pr.ID, oldReviewerID, newRevID)
		if err != nil {
			return err
		}

		revsIDs, err := repos.PR.ListReviewers(ctx, pr.ID)
		if err != nil {
			return err
		}

		pr.Reviewers = revsIDs
		result = pr
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return result, newRevID, nil
}
