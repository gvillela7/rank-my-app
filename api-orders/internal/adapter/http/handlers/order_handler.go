package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.uber.org/zap"
)

type OrderHandler struct {
	useCase   ports.OrderUseCase
	validator *validator.Validate
	logger    *zap.Logger
}

func NewOrderHandler(useCase ports.OrderUseCase, validator *validator.Validate, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		useCase:   useCase,
		validator: validator,
		logger:    logger,
	}
}

// CreateOrder godoc
// @Summary      Create a new order
// @Description  Creates a new order with the provided items
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        order  body      dto.CreateOrderRequest  true  "Order information"
// @Success      201    {object}  SuccessResponseDoc{data=dto.OrderResponse}  "Order created successfully"
// @Failure      400    {object}  ErrorResponseDoc  "Invalid request body or validation error"
// @Failure      500    {object}  ErrorResponseDoc  "Internal server error"
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		ValidationErrorResponse(c, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		ValidationErrorResponse(c, err)
		return
	}

	order, err := h.useCase.CreateOrder(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))

		if httpErr, ok := GetHTTPError(err); ok {
			ErrorResponse(c, httpErr.Code, httpErr, httpErr.Message)
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, err, "Failed to create order")
		return
	}

	SuccessResponse(c, http.StatusCreated, order, "Order created successfully")
}

// GetOrderByID godoc
// @Summary      Get order by ID
// @Description  Retrieves an order by its MongoDB ObjectID
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Order ID (MongoDB ObjectID)"
// @Success      200  {object}  SuccessResponseDoc{data=dto.OrderResponse}  "Order retrieved successfully"
// @Failure      404  {object}  ErrorResponseDoc  "Order not found"
// @Failure      500  {object}  ErrorResponseDoc  "Internal server error"
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.useCase.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get order", zap.Error(err), zap.String("id", id))

		if httpErr, ok := GetHTTPError(err); ok {
			ErrorResponse(c, httpErr.Code, httpErr, httpErr.Message)
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, err, "Failed to get order")
		return
	}

	SuccessResponse(c, http.StatusOK, order, "Order retrieved successfully")
}

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Description  Updates the status of an existing order
// @Tags         Orders
// @Accept       json
// @Produce      json
// @Param        id      path      string                        true  "Order ID (MongoDB ObjectID)"
// @Param        status  body      dto.UpdateOrderStatusRequest  true  "New status"
// @Success      200     {object}  SuccessResponseDoc{data=dto.OrderResponse}  "Order status updated successfully"
// @Failure      400     {object}  ErrorResponseDoc  "Invalid request body or validation error"
// @Failure      404     {object}  ErrorResponseDoc  "Order not found"
// @Failure      500     {object}  ErrorResponseDoc  "Internal server error"
// @Router       /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateOrderStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		ValidationErrorResponse(c, err)
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		ValidationErrorResponse(c, err)
		return
	}

	order, err := h.useCase.UpdateOrderStatus(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update order status", zap.Error(err), zap.String("id", id))

		if httpErr, ok := GetHTTPError(err); ok {
			ErrorResponse(c, httpErr.Code, httpErr, httpErr.Message)
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, err, "Failed to update order status")
		return
	}

	SuccessResponse(c, http.StatusOK, order, "Order status updated successfully")
}
