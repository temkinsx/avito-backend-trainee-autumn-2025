package handler

import (
	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserUsecase domain.UserUsecase
}

func (uh *UserHandler) SetIsActive(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.UsersSetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	userID, active := req.UserID, req.IsActive

	user, err := uh.UserUsecase.SetIsActive(ctx, userID, active)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NOT_FOUND",
					Message: fmt.Sprint(err),
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

	resp := dto.UsersSetIsActiveResponse{User: dto.UserDTO{
		UserID:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}}

	c.JSON(http.StatusOK, resp)
}

func (uh *UserHandler) GetReview(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "user_id is required",
			},
		})
		return
	}

	prs, err := uh.UserUsecase.GetReview(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "NOT_FOUND",
					Message: fmt.Sprint(err),
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

	prsShort := dto.ToPullRequestShortDTOs(prs)

	resp := dto.UsersGetReviewResponse{
		UserID:       userID,
		PullRequests: prsShort,
	}

	c.JSON(http.StatusOK, resp)
}
