package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gvillela7/rank-my-app/internal/infra/rabbitmq"
)

type HealthHandler struct {
	rabbitConn *rabbitmq.RabbitMQConnection
}

func NewHealthHandler(rabbitConn *rabbitmq.RabbitMQConnection) *HealthHandler {
	return &HealthHandler{
		rabbitConn: rabbitConn,
	}
}

// HealthCheck godoc
// @Summary Health check
// @Description Returns the health status of the API and RabbitMQ connection
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx := context.Background()

	apiStatus := "ok"

	rabbitmqStatus := "falha"
	if h.rabbitConn.IsConnected(ctx) {
		rabbitmqStatus = "sucesso"
	}

	c.JSON(http.StatusOK, gin.H{
		"api_status":      apiStatus,
		"rabbitmq_status": rabbitmqStatus,
	})
}
