package main

import (
	"log"

	"financial-api/internal/config"
	"financial-api/internal/handler"
	"financial-api/internal/repository"
	"financial-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := pgxpool.Connect(cfg.Context, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Используем PgRepository (реализацию интерфейса)
	repo := repository.NewPgRepository(pool)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	r := gin.Default()

	r.POST("/topup", h.TopUpBalance)
	r.POST("/transfer", h.TransferMoney)
	r.GET("/transactions/:userID", h.GetLastTransactions)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
