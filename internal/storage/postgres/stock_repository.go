package postgres

import (
	"context"
	"errors"
	"stockwatch/internal/warehouse"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockItemRepository struct {
	pool *pgxpool.Pool
}

func NewStockItemRepository(pool *pgxpool.Pool) *StockItemRepository {
	return &StockItemRepository{
		pool: pool,
	}
}

// AddStockItem implements [warehouse.Repository].
func (s *StockItemRepository) AddStockItem(ctx context.Context, item warehouse.StockItem) error {
	_, err := s.pool.Exec(ctx, `
	INSERT INTO stock_items (name, sku, quantity, threshold, unit)
	VALUES ($1,$2,$3,$4,$5);
	`, item.Name, item.SKU, item.Quantity, item.Threshold, item.Unit)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return warehouse.ErrStockItemAlreadyExists
		}
		return err
	}

	return nil
}

// DeleteStockItem implements [warehouse.Repository].
func (s *StockItemRepository) DeleteStockItem(ctx context.Context, sku string) error {
	tag, err := s.pool.Exec(ctx, `
	DELETE FROM stock_items WHERE sku = $1;
	`, sku)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return warehouse.ErrStockItemNotFound
	}

	return nil
}

// GetStockItem implements [warehouse.Repository].
func (s *StockItemRepository) GetStockItem(ctx context.Context, id int64) (warehouse.StockItem, error) {
	row := s.pool.QueryRow(ctx, `
	SELECT id, name, sku, quantity, threshold, unit, created_at, updated_at 
	FROM stock_items WHERE id = $1;
	`, id)
	var storeItem warehouse.StockItem
	err := row.Scan(&storeItem.ID, &storeItem.Name, &storeItem.SKU, &storeItem.Quantity, &storeItem.Threshold, &storeItem.Unit, &storeItem.CreatedAt, &storeItem.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return warehouse.StockItem{}, warehouse.ErrStockItemNotFound
	}
	if err != nil {
		return warehouse.StockItem{}, err
	}

	return storeItem, nil
}

// GetStockItemBySKU implements [warehouse.Repository].
func (s *StockItemRepository) GetStockItemBySKU(ctx context.Context, sku string) (warehouse.StockItem, error) {
	row := s.pool.QueryRow(ctx, `
	SELECT id, name, sku, quantity, threshold, unit, created_at, updated_at 
	FROM stock_items WHERE sku = $1;
	`, sku)
	var storeItem warehouse.StockItem
	err := row.Scan(&storeItem.ID, &storeItem.Name, &storeItem.SKU, &storeItem.Quantity, &storeItem.Threshold, &storeItem.Unit, &storeItem.CreatedAt, &storeItem.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return warehouse.StockItem{}, warehouse.ErrStockItemNotFound
	}
	if err != nil {
		return warehouse.StockItem{}, err
	}

	return storeItem, nil
}

// ListStockItems implements [warehouse.Repository].
func (s *StockItemRepository) ListStockItems(ctx context.Context) ([]warehouse.StockItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, sku, quantity, threshold, unit, created_at, updated_at 
		FROM stock_items;
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]warehouse.StockItem, 0)

	for rows.Next() {
		var stockItem warehouse.StockItem
		if err := rows.Scan(&stockItem.ID, &stockItem.Name, &stockItem.SKU, &stockItem.Quantity, &stockItem.Threshold, &stockItem.Unit, &stockItem.CreatedAt, &stockItem.UpdatedAt); err != nil {
			return nil, err
		}

		result = append(result, stockItem)
	}

	return result, nil
}

// SetQuantity implements [warehouse.Repository].
func (s *StockItemRepository) SetQuantity(ctx context.Context, sku string, quantity int) (warehouse.StockItem, error) {
	row := s.pool.QueryRow(ctx, `
	UPDATE stock_items
	SET quantity = $1, updated_at = now()
	WHERE sku = $2
	RETURNING id, name, sku, quantity, threshold, unit, created_at, updated_at;
	`, quantity, sku)

	var stockItem warehouse.StockItem

	err := row.Scan(&stockItem.ID, &stockItem.Name, &stockItem.SKU, &stockItem.Quantity, &stockItem.Threshold, &stockItem.Unit, &stockItem.CreatedAt, &stockItem.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return warehouse.StockItem{}, warehouse.ErrStockItemNotFound
	}

	if err != nil {
		return warehouse.StockItem{}, err
	}

	return stockItem, nil

}

var _ warehouse.Repository = (*StockItemRepository)(nil)
