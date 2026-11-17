package notifier

import (
	"context"
	"log"
)

type Log struct{}

func NewLog() *Log { return &Log{} }

func (l *Log) Send(ctx context.Context, text string) error {
	log.Printf("[алерт, заглушка] %s", text)
	return nil
}
