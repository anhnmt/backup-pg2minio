package cmd

import (
	"fmt"
	"strings"
	"sync/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type tele struct {
	enabled  bool
	database string
	chatId   int64
	bot      *tgbotapi.BotAPI
}

var defaultTelegram atomic.Value

func init() {
	SetDefault(&tele{})
}

func Default() *tele {
	return defaultTelegram.Load().(*tele)
}

func SetDefault(t *tele) {
	defaultTelegram.Store(t)
}

func NewTelegram(cfg Telegram, dbName string) (*tele, error) {
	if cfg.Enable == true {
		if cfg.Token == "" {
			return nil, fmt.Errorf("You need to set the %s environment variable", "TELEGRAM_TOKEN")
		}

		if cfg.ChatId == 0 {
			return nil, fmt.Errorf("You need to set the %s environment variable", "TELEGRAM_CHAT_ID")
		}
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	t := &tele{
		enabled:  cfg.Enable,
		database: dbName,
		chatId:   cfg.ChatId,
		bot:      bot,
	}

	return t, nil
}

func OK(text string, a ...any) error {
	log.Info().Msgf(text, a...)
	return Default().Msg(nil, text, a...)
}

func Err(err error, text string, a ...any) error {
	log.Err(err).Msgf(text, a...)
	return Default().Msg(err, text, a...)
}

func (t *tele) Msg(err error, text string, a ...any) error {
	if t == nil {
		return nil
	}

	if !t.enabled || t.chatId == 0 {
		return nil
	}

	if len(a) > 0 {
		text = fmt.Sprintf(text, a...)
	}

	if err != nil {
		text = fmt.Sprintf("\n- Error: %v\n%s", err, text)
	}

	if t.database != "" {
		text = fmt.Sprintf("[%s] - %s", strings.ToUpper(t.database), text)
	}

	if err != nil {
		text = fmt.Sprintf("🔴 %s", text)
	} else {
		text = fmt.Sprintf("🟢 %s", text)
	}

	msg := tgbotapi.NewMessage(t.chatId, text)
	_, sendErr := t.bot.Send(msg)
	if sendErr != nil {
		return sendErr
	}

	return nil
}
