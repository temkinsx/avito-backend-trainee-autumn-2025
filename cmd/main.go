package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"avito-backend-trainee-autumn-2025/internal/api/handler"
	"avito-backend-trainee-autumn-2025/internal/api/route"
	"avito-backend-trainee-autumn-2025/internal/repository/postgres"
	"avito-backend-trainee-autumn-2025/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	txManager := postgres.NewTxManager(pool)

	userRepo := postgres.NewUserRepository(pool)
	teamRepo := postgres.NewTeamRepository(pool)
	prRepo := postgres.NewPRRepository(pool)

	userUC := usecase.NewUserUsecase(userRepo, prRepo, txManager)
	teamUC := usecase.NewTeamUsecase(teamRepo, userRepo, txManager)
	prUC := usecase.NewPRUsecase(userRepo, prRepo, txManager)

	userHandler := &handler.UserHandler{UserUsecase: userUC}
	teamHandler := &handler.TeamHandler{TeamUsecase: teamUC}
	prHandler := &handler.PRHandler{PRUsecase: prUC}

	router := gin.Default()
	route.Register(router, prHandler, teamHandler, userHandler)

	serverErr := make(chan error, 1)
	go func() {
		if err := router.Run(":8080"); err != nil {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		stop()
	case err := <-serverErr:
		if err != nil {
			log.Printf("server stopped: %v", err)
		}
	}

	os.Exit(0)
}
