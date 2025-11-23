package handler

import (
	"avito-backend-trainee-autumn-2025/internal/api/dto"
	"avito-backend-trainee-autumn-2025/internal/domain"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	TeamUsecase domain.TeamUsecase
}

func (th *TeamHandler) Add(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.TeamAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "BAD_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	team := &domain.Team{
		Name:    req.TeamName,
		Members: make([]*domain.User, 0, len(req.Members)),
	}

	for _, m := range req.Members {
		team.Members = append(team.Members, &domain.User{
			ID:       m.UserID,
			Name:     m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	createdTeam, err := th.TeamUsecase.Add(ctx, team)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
				Error: dto.ErrorDTO{
					Code:    "TEAM_EXISTS",
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

	resp := dto.TeamAddResponse{
		Team: dto.ToTeamDTO(createdTeam),
	}

	c.JSON(http.StatusCreated, resp)
}

func (th *TeamHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()

	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDTO{
				Code:    "VALIDATION_ERROR",
				Message: "team_name is required",
			},
		})
		return
	}

	team, err := th.TeamUsecase.ListByName(ctx, teamName)
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

	c.JSON(http.StatusOK, dto.ToTeamDTO(team))
}
