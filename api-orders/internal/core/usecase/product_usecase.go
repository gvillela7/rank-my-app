package usecase

import (
	"context"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/ports"
)

type productUseCase struct {
	repository ports.ProductRepository
}

func NewProductUseCase(repository ports.ProductRepository) ports.ProductUseCase {
	return &productUseCase{
		repository: repository,
	}
}

func (uc *productUseCase) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := &domain.Product{
		Name:        req.Name,
		Description: req.Description,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}

	if err := uc.repository.Create(ctx, product); err != nil {
		return nil, err
	}

	return dto.ToProductResponse(
		product.ID,
		product.Name,
		product.Description,
		product.Quantity,
		product.Price,
		product.CreatedAt,
		product.UpdatedAt,
	), nil
}
