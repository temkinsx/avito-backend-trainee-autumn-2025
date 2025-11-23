package domain

import "errors"

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrPRMerged      = errors.New("cannot reassign on merged PR")
	ErrNotAssigned   = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate   = errors.New("no active replacement candidate in team")
	ErrNotFound      = errors.New("not found")
)
