package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
	"go.uber.org/zap"
)

type ProductHandler struct {
	useCase   ports.ProductUseCase
	validator *validator.Validate
	logger    *zap.Logger
}

func NewProductHandler(useCase ports.ProductUseCase, validator *validator.Validate, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		useCase:   useCase,
		validator: validator,
		logger:    logger,
	}
}

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Creates a new product with the provided information
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body      dto.CreateProductRequest  true  "Product information"
// @Success      201      {object}  SuccessResponseDoc{data=dto.ProductResponse}  "Product created successfully"
// @Failure      400      {object}  ErrorResponseDoc  "Invalid request body or validation error"
// @Failure      500      {object}  ErrorResponseDoc  "Internal server error"
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest

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

	product, err := h.useCase.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))

		if httpErr, ok := GetHTTPError(err); ok {
			ErrorResponse(c, httpErr.Code, httpErr, httpErr.Message)
			return
		}

		ErrorResponse(c, http.StatusInternalServerError, err, "Failed to create product")
		return
	}

	SuccessResponse(c, http.StatusCreated, product, "Product created successfully")
}
