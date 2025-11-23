package route

import (
	"net/http"

	"avito-backend-trainee-autumn-2025/internal/api/handler"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, prHandler *handler.PRHandler, teamHandler *handler.TeamHandler, userHandler *handler.UserHandler) {
	team := r.Group("/team")
	{
		team.POST("/add", teamHandler.Add)
		team.GET("/get", teamHandler.Get)
	}

	pr := r.Group("/pullRequest")
	{
		pr.POST("/create", prHandler.Create)
		pr.POST("/merge", prHandler.Merge)
		pr.POST("/reassign", prHandler.Reassign)
	}

	users := r.Group("/users")
	{
		users.POST("/setIsActive", userHandler.SetIsActive)
		users.GET("/getReview", userHandler.GetReview)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
