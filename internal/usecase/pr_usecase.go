package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
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
		createdPR, err := repos.PR.Create(ctx, newPR)
		if err != nil {
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
	return p.prRepository.UpdateStatusMerged(ctx, prID)
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

		author, err := repos.User.FetchByID(ctx, pr.AuthorID)
		if err != nil {
			return err
		}

		revs, err := repos.User.FetchActiveByTeam(ctx, author.TeamName, author.ID, oldReviewerID)
		if err != nil {
			return err
		}
		if len(revs) == 0 {
			return domain.ErrNoCandidate
		}

		newRevID = revs[0].ID

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
