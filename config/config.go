package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Configs struct {
	// подключения и секреты (из .env)
	DatabaseURL    string
	TelegramToken  string
	TelegramChatID int64

	// поведение мониторинга
	PollInterval   time.Duration // как часто опрашивать принтеры, 30m
	TonerThreshold int           // 15
	AlertCooldown  time.Duration // 24h
	MaxConcurrent  int           // сколько принтеров опрашивать одновременно, 10
	PollTimeout    time.Duration // дедлайн на один принтер, 15s
}

// Load читает настройки из окружения. Файл .env необязателен:
// на сервере переменные могут быть заданы напрямую.
// Уже заданные переменные окружения имеют приоритет над .env.
func Load() (Configs, error) {
	if err := godotenv.Load("../../.env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Configs{}, fmt.Errorf("чтение .env: %w", err)
	}

	var l loader

	cfg := Configs{
		DatabaseURL:    l.str("DATABASE_URL", ""),
		TelegramToken:  l.str("TELEGRAM_TOKEN", ""),
		TelegramChatID: l.int64("TELEGRAM_CHAT_ID", 0),

		PollInterval:   l.duration("POLL_INTERVAL", 30*time.Minute),
		TonerThreshold: l.int("TONER_THRESHOLD", 15),
		AlertCooldown:  l.duration("ALERT_COOLDOWN", 24*time.Hour),
		MaxConcurrent:  l.int("MAX_CONCURRENT", 10),
		PollTimeout:    l.duration("POLL_TIMEOUT", 15*time.Second),
	}

	errs := append(l.errs, cfg.validate()...)
	if err := errors.Join(errs...); err != nil {
		return Configs{}, fmt.Errorf("некорректная конфигурация:\n%w", err)
	}

	return cfg, nil
}

func (c Configs) validate() []error {
	var errs []error

	if c.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL обязателен"))
	}
	if c.TelegramToken != "" && c.TelegramChatID == 0 {
		errs = append(errs, errors.New("TELEGRAM_CHAT_ID обязателен, если задан TELEGRAM_TOKEN"))
	}
	if c.TonerThreshold < 1 || c.TonerThreshold > 100 {
		errs = append(errs, fmt.Errorf("TONER_THRESHOLD должен быть от 1 до 100, получено %d", c.TonerThreshold))
	}
	if c.MaxConcurrent <= 0 {
		errs = append(errs, fmt.Errorf("MAX_CONCURRENT должен быть больше 0, получено %d", c.MaxConcurrent))
	}
	if c.PollInterval <= 0 {
		errs = append(errs, fmt.Errorf("POLL_INTERVAL должен быть больше 0, получено %v", c.PollInterval))
	}
	if c.AlertCooldown <= 0 {
		errs = append(errs, fmt.Errorf("ALERT_COOLDOWN должен быть больше 0, получено %v", c.AlertCooldown))
	}
	if c.PollTimeout <= 0 {
		errs = append(errs, fmt.Errorf("POLL_TIMEOUT должен быть больше 0, получено %v", c.PollTimeout))
	}
	if c.PollInterval > 0 && c.AlertCooldown > 0 && c.PollInterval > c.AlertCooldown {
		errs = append(errs, errors.New("POLL_INTERVAL не должен превышать ALERT_COOLDOWN"))
	}

	return errs
}

// loader читает переменные окружения и копит ошибки разбора,
// чтобы показать все проблемы сразу, а не по одной.
type loader struct {
	errs []error
}

func (l *loader) str(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (l *loader) int(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: %q не является числом", key, v))
		return def
	}
	return n
}

func (l *loader) int64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: %q не является числом", key, v))
		return def
	}
	return n
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: %q не является длительностью (примеры: 30s, 15m, 24h)", key, v))
		return def
	}
	return d
}
