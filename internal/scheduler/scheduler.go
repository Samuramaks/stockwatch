package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Schedule struct {
	Checker  Checker
	Interval time.Duration
}

func NewSchedule(checker Checker, interval time.Duration) (*Schedule, error) {
	if interval <= 0 {

		return nil, fmt.Errorf("Интервал не может быть нулевым или отрицательным значением: %v", interval)
	}

	return &Schedule{
		Checker:  checker,
		Interval: interval,
	}, nil

}

func (s *Schedule) Run(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	s.check(ctx)

	ticker := time.NewTicker(s.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// s.check(ctx)
			log.Print("Таймер вышел")
		case <-ctx.Done():
			return
		}
	}

}

func (s *Schedule) check(ctx context.Context) {
	if err := s.Checker.CheckAll(ctx); err != nil && ctx.Err() == nil {
		log.Printf("проверка не удалась: %v", err)
	}
}
