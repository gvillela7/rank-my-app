package ports

import (
	"context"

	"github.com/gvillela7/rank-my-app/internal/core/dto"
)

type OrderUseCase interface {
	ProcessOrderStatusMessage(ctx context.Context, message *dto.OrderStatusMessage) error
}
