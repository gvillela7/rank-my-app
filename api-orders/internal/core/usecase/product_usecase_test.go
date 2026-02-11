package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gvillela7/rank-my-app/internal/core/domain"
	"github.com/gvillela7/rank-my-app/internal/core/dto"
	"github.com/gvillela7/rank-my-app/internal/core/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Mock Repository
type mockProductRepository struct {
	createFunc   func(ctx context.Context, product *domain.Product) error
	findByIDFunc func(ctx context.Context, id primitive.ObjectID) (*domain.Product, error)
}

func (m *mockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, product)
	}
	return nil
}

func (m *mockProductRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func TestProductUseCase_CreateProduct_Success(t *testing.T) {
	mockRepo := &mockProductRepository{
		createFunc: func(ctx context.Context, product *domain.Product) error {
			product.ID = primitive.NewObjectID()
			return nil
		},
	}

	uc := usecase.NewProductUseCase(mockRepo)

	req := &dto.CreateProductRequest{
		Name:        "Test Product",
		Description: "Test Description",
		Quantity:    10,
		Price:       99.99,
	}

	resp, err := uc.CreateProduct(context.Background(), req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}

	if resp.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, resp.Name)
	}

	if resp.Price != req.Price {
		t.Errorf("Expected price %f, got %f", req.Price, resp.Price)
	}
}

func TestProductUseCase_CreateProduct_RepositoryError(t *testing.T) {
	expectedErr := errors.New("database connection failed")

	mockRepo := &mockProductRepository{
		createFunc: func(ctx context.Context, product *domain.Product) error {
			return expectedErr
		},
	}

	uc := usecase.NewProductUseCase(mockRepo)

	req := &dto.CreateProductRequest{
		Name:        "Test Product",
		Description: "Test Description",
		Quantity:    10,
		Price:       99.99,
	}

	resp, err := uc.CreateProduct(context.Background(), req)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	if resp != nil {
		t.Error("Expected nil response on error")
	}
}
