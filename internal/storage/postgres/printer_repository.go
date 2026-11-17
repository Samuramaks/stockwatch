package postgres

import (
	"context"
	"errors"
	"stockwatch/internal/printer"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrinterRepository struct {
	pool *pgxpool.Pool
}

func NewPrinterRepository(pool *pgxpool.Pool) *PrinterRepository {
	return &PrinterRepository{
		pool: pool,
	}
}

// UpsertStatus implements [printer.StatusRepository].
func (p *PrinterRepository) UpsertStatus(ctx context.Context, id int64, levelPct int, lastPolledAt time.Time) error {
	_, err := p.pool.Exec(ctx, `
	INSERT INTO printer_statuses (printer_id, level_pct, last_polled_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (printer_id) DO UPDATE
	SET level_pct = EXCLUDED.level_pct,
	    last_polled_at = EXCLUDED.last_polled_at;
	`, id, levelPct, lastPolledAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return printer.ErrPrinterNotFound
		}
		return err
	}

	return nil
}

// RecordAlertSent implements [printer.StatusRepository].
func (p *PrinterRepository) RecordAlertSent(ctx context.Context, id int64, lastAlertedAt time.Time) error {
	return p.setAlertedAt(ctx, id, &lastAlertedAt)
}

// ClearAlert implements [printer.StatusRepository].
func (p *PrinterRepository) ClearAlert(ctx context.Context, id int64) error {
	return p.setAlertedAt(ctx, id, nil)
}

// setAlertedAt записывает last_alerted_at; nil сбрасывает значение в NULL.
func (p *PrinterRepository) setAlertedAt(ctx context.Context, id int64, alertedAt *time.Time) error {
	tag, err := p.pool.Exec(ctx, `
	UPDATE printer_statuses
	SET last_alerted_at = $1
	WHERE printer_id = $2;
	`, alertedAt, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return printer.ErrPrinterNotFound
	}

	return nil
}

// AddPrinter implements [printer.Repository].
func (p *PrinterRepository) AddPrinter(ctx context.Context, printerItem printer.Printers) error {
	_, err := p.pool.Exec(ctx, `
	INSERT INTO printers (name, ip, model, location, cartridge_sku, snmp_community)
	VALUES ($1, $2, $3, $4, $5, $6);
	`, printerItem.Name, printerItem.IP, printerItem.Model, printerItem.Location, printerItem.CartridgeSKU, printerItem.SNMPCommunity)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return printer.ErrPrinterAlreadyExists
		}
		return err
	}

	return nil
}

// DeletePrinter implements [printer.Repository].
func (p *PrinterRepository) DeletePrinter(ctx context.Context, id int64) error {
	tag, err := p.pool.Exec(ctx, `
	DELETE FROM printers WHERE id = $1;
	`, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return printer.ErrPrinterNotFound

	}
	return nil
}

// GetPrinter implements [printer.Repository].
func (p *PrinterRepository) GetPrinter(ctx context.Context, id int64) (printer.Printers, error) {
	row := p.pool.QueryRow(ctx, `
	SELECT id, name, ip, model, location, cartridge_sku, snmp_community
	FROM printers WHERE id = $1;
	`, id)
	var printerItem printer.Printers
	err := row.Scan(&printerItem.ID, &printerItem.Name, &printerItem.IP, &printerItem.Model, &printerItem.Location, &printerItem.CartridgeSKU, &printerItem.SNMPCommunity)
	if errors.Is(err, pgx.ErrNoRows) {
		return printer.Printers{}, printer.ErrPrinterNotFound
	}
	if err != nil {
		return printer.Printers{}, err
	}

	return printerItem, nil
}

// ListPrintersWithStatus implements [printer.Repository].
func (p *PrinterRepository) ListPrintersWithStatus(ctx context.Context) ([]printer.Printers, error) {
	rows, err := p.pool.Query(ctx, `
	SELECT p.id, p.name, p.ip, p.model, p.location, p.cartridge_sku, p.snmp_community,
       s.level_pct, s.last_polled_at, s.last_alerted_at
	FROM printers p
	LEFT JOIN printer_statuses s ON s.printer_id = p.id
	ORDER BY p.name;
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]printer.Printers, 0)

	for rows.Next() {
		var printerItem printer.Printers
		var levelPct *int
		var polledAt, alertedAt *time.Time

		if err := rows.Scan(
			&printerItem.ID,
			&printerItem.Name,
			&printerItem.IP,
			&printerItem.Model,
			&printerItem.Location,
			&printerItem.CartridgeSKU,
			&printerItem.SNMPCommunity,
			&levelPct,
			&polledAt,
			&alertedAt,
		); err != nil {
			return nil, err
		}

		if levelPct != nil && polledAt != nil {
			printerItem.Status = &printer.PrinterStatus{
				LevelPct:      *levelPct,
				LastPolledAt:  *polledAt,
				LastAlertedAt: alertedAt,
			}
		}

		result = append(result, printerItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

var _ printer.Repository = (*PrinterRepository)(nil)
var _ printer.StatusRepository = (*PrinterRepository)(nil)
