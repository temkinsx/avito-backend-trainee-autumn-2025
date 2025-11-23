package dto

import (
	"avito-backend-trainee-autumn-2025/internal/domain"
	"time"
)

type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type PullRequestCreateRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PullRequestCreateResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type PullRequestMergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type PullRequestMergeResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type PullRequestReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type PullRequestReassignResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

func ToPullRequestDTO(pr *domain.PullRequest) PullRequestDTO {
	if pr == nil {
		return PullRequestDTO{}
	}

	return PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: append([]string(nil), pr.Reviewers...),
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func ToPullRequestShortDTO(pr *domain.PullRequest) PullRequestShortDTO {
	if pr == nil {
		return PullRequestShortDTO{}
	}

	return PullRequestShortDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}

func ToPullRequestDTOs(prs []*domain.PullRequest) []PullRequestDTO {
	res := make([]PullRequestDTO, 0, len(prs))
	for _, pr := range prs {
		res = append(res, ToPullRequestDTO(pr))
	}
	return res
}

func ToPullRequestShortDTOs(prs []*domain.PullRequest) []PullRequestShortDTO {
	res := make([]PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		res = append(res, ToPullRequestShortDTO(pr))
	}
	return res
}

func ToPullRequestCreateResponse(pr *domain.PullRequest) PullRequestCreateResponse {
	return PullRequestCreateResponse{
		PR: ToPullRequestDTO(pr),
	}
}

func ToPullRequestMergeResponse(pr *domain.PullRequest) PullRequestMergeResponse {
	return PullRequestMergeResponse{
		PR: ToPullRequestDTO(pr),
	}
}

func ToPullRequestReassignResponse(pr *domain.PullRequest, replacedBy string) PullRequestReassignResponse {
	return PullRequestReassignResponse{
		PR:         ToPullRequestDTO(pr),
		ReplacedBy: replacedBy,
	}
}
