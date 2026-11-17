package notifier

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Telegram struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegram(token string, chatID int64) (*Telegram, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	bot, err := tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil {
		return nil, fmt.Errorf("telegram send: %w", err)
	}

	// bot.Debug = debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	telegram := &Telegram{
		bot:    bot,
		chatID: chatID,
	}

	return telegram, nil
}

func (t *Telegram) Send(ctx context.Context, text string) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	msg := tgbotapi.NewMessage(t.chatID, text)
	_, err := t.bot.Send(msg)

	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}

	return nil

}
