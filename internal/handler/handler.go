package handler

import (
	"financial-api/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// Запрос для пополнения
type TopUpRequest struct {
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
}

// Запрос для перевода
type TransferRequest struct {
	FromUserID int     `json:"from_user_id"`
	ToUserID   int     `json:"to_user_id"`
	Amount     float64 `json:"amount"`
}

// POST /topup
func (h *Handler) TopUpBalance(c *gin.Context) {
	var req TopUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.service.TopUpBalance(c, req.UserID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "balance topped up"})
}

// POST /transfer
func (h *Handler) TransferMoney(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.service.TransferMoney(c, req.FromUserID, req.ToUserID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer success"})
}

// GET /transactions/:userID
func (h *Handler) GetLastTransactions(c *gin.Context) {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userID"})
		return
	}

	transactions, err := h.service.GetLastTransactions(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
