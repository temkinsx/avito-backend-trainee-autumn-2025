package usecase

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"context"
)

type userUsecase struct {
	userRepository domain.UserRepository
	prRepository   domain.PRRepository
	txManager      domain.TxManager
}

func NewUserUsercase(userRepository domain.UserRepository, prRepository domain.PRRepository, txManager domain.TxManager) domain.UserUsecase {
	return &userUsecase{
		userRepository: userRepository,
		prRepository:   prRepository,
		txManager:      txManager,
	}
}

func (u *userUsecase) SetIsActive(ctx context.Context, userID string, active bool) (*domain.User, error) {
	return u.userRepository.UpdateIsActive(ctx, userID, active)
}

func (u *userUsecase) GetReview(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	exists, err := u.userRepository.Exists(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, domain.ErrNotFound
	}

	return u.prRepository.ListReviewableByUserID(ctx, userID)
}
