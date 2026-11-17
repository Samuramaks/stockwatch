package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"stockwatch/internal/https"
	"stockwatch/internal/monitoring"
	"stockwatch/internal/notifier"
	"stockwatch/internal/scheduler"
	"stockwatch/internal/snmp"
	"stockwatch/internal/storage/postgres"
	"sync"
	"syscall"

	"stockwatch/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("ошибка: %v", err)
	}
}

func run() error {
	//контекст отменяется по Ctrl+C или SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load()

	if err != nil {
		return fmt.Errorf("конфиг: %w", err)
	}

	//---хранилище---
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("подключение к БД: %w", err)
	}

	defer pool.Close()

	stockRepo := postgres.NewStockItemRepository(pool)
	printerRepo := postgres.NewPrinterRepository(pool)

	//---уведомления---
	var alerts monitoring.Notifier
	if cfg.TelegramToken == "" {
		log.Println("TELEGRAM_TOKEN не задан: алерты будут только в логе")
		alerts = notifier.NewLog()
	} else {
		tg, err := notifier.NewTelegram(cfg.TelegramToken, cfg.TelegramChatID)
		if err != nil {
			return fmt.Errorf("telegram: %w", err)
		}
		alerts = tg
	}

	//---мониторинг---
	mon, err := monitoring.NewMonitor(printerRepo, snmp.NewClient(), stockRepo, alerts, cfg)
	if err != nil {
		return fmt.Errorf("monitor: %w", err)
	}

	sched, err := scheduler.NewSchedule(mon, cfg.PollTimeout)
	if err != nil {
		return fmt.Errorf("scheduler: %w", err)
	}

	// ---HTTP---

	server := https.NewHTTPServer(https.NewHTTPHandler(stockRepo), https.NewPrinterHandlers(printerRepo))

	//---запуск---
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		sched.Run(ctx)
	}()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.StartServer()
	}()
	select {
	case <-ctx.Done():
		log.Println("Получен сигнал остановки")
	case err := <-serverErr:
		stop()
		wg.Wait()
		return fmt.Errorf("http-сервер: %w", err)
	}

	wg.Wait()
	log.Println("Остановлено")
	return nil

}
