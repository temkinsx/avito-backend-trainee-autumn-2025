package handler

import (
	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PRHandler struct {
	PRUsecase domain.PRUsecase
}

func (h *PRHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.PullRequestCreateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	pr := &domain.PullRequest{
		ID:       req.PullRequestID,
		Name:     req.PullRequestName,
		AuthorID: req.AuthorID,
		Status:   domain.StatusOpen,
	}

	createdPR, err := h.PRUsecase.CreateWithReviewers(ctx, pr)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		case errors.Is(err, domain.ErrAlreadyExists):
			c.JSON(http.StatusConflict, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "PR_EXISTS",
					Message: err.Error(),
				},
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "INTERNAL_ERROR",
					Message: err.Error(),
				},
			})
			return
		}
	}

	resp := dto.ToPullRequestCreateResponse(createdPR)

	c.JSON(http.StatusCreated, resp)
}

func (h *PRHandler) Merge(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.PullRequestMergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	pr, err := h.PRUsecase.Merge(ctx, req.PullRequestID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	resp := dto.ToPullRequestMergeResponse(pr)

	c.JSON(http.StatusOK, resp)
}

func (h *PRHandler) Reassign(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.PullRequestReassignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	pr, newRevID, err := h.PRUsecase.Reassign(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		case errors.Is(err, domain.ErrPRMerged):
			c.JSON(http.StatusConflict, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "PR_MERGED",
					Message: err.Error(),
				},
			})
			return
		case errors.Is(err, domain.ErrNotAssigned):
			c.JSON(http.StatusConflict, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NOT_ASSIGNED",
					Message: err.Error(),
				},
			})
			return
		case errors.Is(err, domain.ErrNoCandidate):
			c.JSON(http.StatusConflict, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NO_CANDIDATE",
					Message: err.Error(),
				},
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "INTERNAL_ERROR",
					Message: err.Error(),
				},
			})
			return
		}
	}

	resp := dto.ToPullRequestReassignResponse(pr, newRevID)

	c.JSON(http.StatusOK, resp)
}
