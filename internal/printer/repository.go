package printer

import (
	"context"
	"time"
)

type Repository interface {
	AddPrinter(ctx context.Context, printer Printers) error
	GetPrinter(ctx context.Context, id int64) (Printers, error)
	ListPrintersWithStatus(ctx context.Context) ([]Printers, error)
	DeletePrinter(ctx context.Context, id int64) error
}

type StatusRepository interface {
	ListPrintersWithStatus(ctx context.Context) ([]Printers, error)
	UpsertStatus(ctx context.Context, id int64, levelPct int, lastPolledAt time.Time) error
	RecordAlertSent(ctx context.Context, id int64, lastAlertedAt time.Time) error
	ClearAlert(ctx context.Context, id int64) error
}
