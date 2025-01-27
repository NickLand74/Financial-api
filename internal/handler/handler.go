package handler

import (
	"financial-api/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) TopUpBalance(c *gin.Context) {
	// Implement logic
}

func (h *Handler) TransferMoney(c *gin.Context) {
	// Implement logic
}

func (h *Handler) GetLastTransactions(c *gin.Context) {
	// Implement logic
}
