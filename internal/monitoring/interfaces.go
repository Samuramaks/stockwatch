package monitoring

import (
	"context"
	"stockwatch/internal/warehouse"
)

type TonerReader interface {
	GetTonerLevel(ctx context.Context, ip, community string) (int, error)
}

type StockReader interface {
	GetStockItemBySKU(ctx context.Context, sku string) (warehouse.StockItem, error)
}

type Notifier interface {
	Send(ctx context.Context, text string) error
}
