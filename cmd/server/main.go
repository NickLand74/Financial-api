// main.go
package main

import (
	"financial-api/internal/config"
	"financial-api/internal/handler"
	"financial-api/internal/repository"
	"financial-api/internal/service"
	"log"

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

	repo := repository.NewRepository(pool) // Убедитесь, что это интерфейс, а не указатель на него
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
