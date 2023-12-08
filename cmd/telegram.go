package cmd

import (
	"fmt"
	"strings"
	"sync/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

type Telegram struct {
	enabled  bool
	database string
	chatId   int64
	bot      *tgbotapi.BotAPI
}

var defaultTelegram atomic.Value

func Default() *Telegram {
	return defaultTelegram.Load().(*Telegram)
}

func SetDefault(t *Telegram) {
	defaultTelegram.Store(t)
}

func NewTelegram() (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(viper.GetString(TelegramToken))
	if err != nil {
		return nil, err
	}

	t := &Telegram{
		enabled:  viper.GetBool(TelegramEnabled),
		database: viper.GetString(PostgresDatabase),
		chatId:   viper.GetInt64(TelegramChatId),
		bot:      bot,
	}

	return t, nil
}

func OK(text string, a ...any) error {
	return Default().Msg(nil, text, a...)
}

func Err(err error, text string, a ...any) error {
	return Default().Msg(err, text, a...)
}

func (t *Telegram) Msg(err error, text string, a ...any) error {
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
