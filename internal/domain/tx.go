package domain

import "context"

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context, repos *Repos) error) error
}

type Repos struct {
	PR   PRRepository
	User UserRepository
	Team TeamRepository
}
