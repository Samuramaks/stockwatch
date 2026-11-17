package monitoring

import (
	"fmt"

	"stockwatch/internal/printer"
	"stockwatch/internal/warehouse"
)

const urgentPrefix = "❗️❗️❗️"

// alertText собирает текст алерта о низком уровне тонера.
// stock == nil означает, что данных по складу нет.
func alertText(p printer.Printers, level int, stock *warehouse.StockItem) string {
	head := fmt.Sprintf("Принтер «%s»", p.Name)
	if p.Location != "" {
		head += fmt.Sprintf(" (%s)", p.Location)
	}
	head += fmt.Sprintf(": тонер %d%%.", level)

	switch {
	case stock == nil:
		return head + " Данные по складу недоступны."
	case stock.Quantity <= 0:
		return fmt.Sprintf("%s %s На складе %s: %s. Нужно закупить картриджи!",
			urgentPrefix, head, stock.SKU, formatQuantity(stock))
	default:
		return fmt.Sprintf("%s На складе %s: %s.", head, stock.SKU, formatQuantity(stock))
	}
}

// formatQuantity выводит количество с единицей измерения.
// Отрицательное значение показываем как 0.
func formatQuantity(s *warehouse.StockItem) string {
	q := s.Quantity
	if q < 0 {
		q = 0
	}
	if s.Unit == "" {
		return fmt.Sprintf("%d", q)
	}
	return fmt.Sprintf("%d %s", q, s.Unit)
}
