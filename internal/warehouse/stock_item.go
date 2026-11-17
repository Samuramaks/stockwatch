package warehouse

import (
	"context"
	"errors"
	"time"
)

// кастомные ошибки связанные со складом
var (
	ErrStockItemNotFound      = errors.New("расходник не найден")
	ErrStockItemAlreadyExists = errors.New("расходник уже существует")
)

// StockItem описывает складскую позицию — расходный материал,
// который хранится на складе (например, картридж под конкретную
// модель принтера). Количество отслеживается вручную через API,
// а порог Threshold используется для определения критично малого
// остатка (см. метод IsLow).
type StockItem struct {
	ID        int64
	Name      string // человекочитаемое название позиции
	SKU       string // уникальный код позиции, используется для поиска и связи с Printer.CartridgeSKU
	Quantity  int    // текущее количество на складе
	Threshold int    // порог, ниже или равный которому остаток считается низким
	Unit      string // единица измерения ("шт", "упаковка" и тд)
	CreatedAt time.Time
	UpdatedAt time.Time // обновляется при каждом изменении Quantity
}

func (s StockItem) IsLow() bool {
	return s.Quantity <= s.Threshold
}

type Repository interface {
	AddStockItem(ctx context.Context, item StockItem) error
	GetStockItem(ctx context.Context, id int64) (StockItem, error)
	GetStockItemBySKU(ctx context.Context, sku string) (StockItem, error)
	ListStockItems(ctx context.Context) ([]StockItem, error)
	SetQuantity(ctx context.Context, sku string, quantity int) (StockItem, error)
	DeleteStockItem(ctx context.Context, sku string) error
}
