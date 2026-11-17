package monitoring

import (
	"context"
	"errors"
	"fmt"
	"log"
	"stockwatch/config"
	"stockwatch/internal/printer"
	"stockwatch/internal/warehouse"
	"sync"
	"time"
)

type Monitor struct {
	printers printer.StatusRepository
	toner    TonerReader
	stock    StockReader
	notifier Notifier
	cfg      config.Configs
}

func NewMonitor(printers printer.StatusRepository, toner TonerReader, stock StockReader, notifier Notifier, cfg config.Configs) (*Monitor, error) {
	if cfg.AlertCooldown <= 0 || cfg.MaxConcurrent <= 0 || cfg.TonerThreshold <= 0 || cfg.PollTimeout <= 0 {
		return nil, fmt.Errorf("Поля: AlertCooldown: %s, MaxConcurrent: %v, TonerThreshold: %v, PoollTimeout: %s", cfg.AlertCooldown, cfg.MaxConcurrent, cfg.TonerThreshold, cfg.PollTimeout)
	}
	return &Monitor{
		printers: printers,
		toner:    toner,
		stock:    stock,
		notifier: notifier,
		cfg:      cfg,
	}, nil
}

type resultWithError struct {
	Printer    printer.Printers
	TonerLevel int
	Survey     bool
	Err        error
}

func shouldAlert(level int, status *printer.PrinterStatus, now time.Time, threshold int, cooldown time.Duration) bool {
	if level > threshold {
		return false
	}

	if status == nil || status.LastAlertedAt == nil {
		return true
	}

	return now.Sub(*status.LastAlertedAt) >= cooldown
}

func pollWorker(ctx context.Context, wg *sync.WaitGroup, toner TonerReader, timeout time.Duration, inCh <-chan printer.Printers, outCh chan<- resultWithError) {
	defer wg.Done()

	for printer := range inCh {
		pollCtx, cancel := context.WithTimeout(ctx, timeout)
		tonerLevel, err := toner.GetTonerLevel(pollCtx, printer.IP, printer.SNMPCommunity)

		cancel()

		select {
		case outCh <- resultWithError{
			Printer:    printer,
			TonerLevel: tonerLevel,
			Survey:     true,
			Err:        err,
		}:
		//
		case <-ctx.Done():
			return
		}

	}
}

func (m *Monitor) pollPrinters(ctx context.Context, printers []printer.Printers) []resultWithError {
	// канал для передачи входных данных горутинам
	inputCh := make(chan printer.Printers)
	// канал для сбора результатов
	outputCh := make(chan resultWithError)

	go func() {
		defer close(inputCh)
		for _, p := range printers {
			select {
			case inputCh <- p:
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < m.cfg.MaxConcurrent; i++ {
		wg.Add(1)
		go pollWorker(ctx, &wg, m.toner, m.cfg.PollTimeout, inputCh, outputCh)
	}

	go func() {
		wg.Wait()
		close(outputCh)
	}()

	results := make([]resultWithError, 0, len(printers))

	for res := range outputCh {

		results = append(results, res)
	}

	return results
}

func (m *Monitor) getItemBySku(ctx context.Context, sku string) *warehouse.StockItem {
	item, err := m.stock.GetStockItemBySKU(ctx, sku)

	if err != nil {
		if !errors.Is(err, warehouse.ErrStockItemNotFound) {
			log.Printf("склад: не удалось получить %q: %v", sku, err)
		}
		return nil
	}
	return &item
}

func (m *Monitor) processPrinter(ctx context.Context, result resultWithError, now time.Time) {
	p := result.Printer

	if result.Err != nil {
		log.Printf("принтер %d (%s): опрос не удался: %v", p.ID, p.Name, result.Err)
		return
	}

	if err := m.printers.UpsertStatus(ctx, result.Printer.ID, result.TonerLevel, now); err != nil {
		log.Printf("принтер %d (%s): не удалось записать статус: %v", p.ID, p.Name, err)
		return
	}

	if result.TonerLevel > m.cfg.TonerThreshold {
		if p.Status != nil && p.Status.LastAlertedAt != nil {
			if err := m.printers.ClearAlert(ctx, p.ID); err != nil {
				log.Printf("принтер %d (%s): не удалось сбросить алерт: %v", p.ID, p.Name, err)
			}

		}
		return
	}

	if !shouldAlert(result.TonerLevel, p.Status, now, m.cfg.TonerThreshold, m.cfg.AlertCooldown) {
		return
	}

	stock := m.getItemBySku(ctx, p.CartridgeSKU)
	text := alertText(p, result.TonerLevel, stock)

	if err := m.notifier.Send(ctx, text); err != nil {
		log.Printf("принтер %d (%s): алерт не отправлен: %v", p.ID, p.Name, err)
		return
	}

	if err := m.printers.RecordAlertSent(ctx, p.ID, now); err != nil {
		log.Printf("принтер %d (%s): алерт отправлен, но не отмечен: %v", p.ID, p.Name, err)
	}

}

func (m *Monitor) CheckAll(ctx context.Context) error {
	printers, err := m.printers.ListPrintersWithStatus(ctx)
	if err != nil {
		return fmt.Errorf("получение списка принтеров: %w", err)
	}

	results := m.pollPrinters(ctx, printers)

	// при отмене результаты могут быть неполными, обрабатывать их не нужно
	if err := ctx.Err(); err != nil {
		return err
	}

	now := time.Now()

	failed := 0
	for _, res := range results {
		if err := ctx.Err(); err != nil {
			return err
		}
		if res.Err != nil {
			failed++
		}
		m.processPrinter(ctx, res, now)
	}

	log.Printf("опрос завершён: принтеров %d, ошибок опроса %d", len(printers), failed)
	return nil

	// log.Print("Метод сбора")
	// return nil
}
